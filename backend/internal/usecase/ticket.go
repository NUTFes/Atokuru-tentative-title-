package usecase

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/repository"
)

// maxCommentLength は備考の最大の文字数（仕様書の maxLength と合わせる）。
const maxCommentLength = 200

// TicketUsecase は整理券についての処理をまとめる。
type TicketUsecase struct {
	stalls  *repository.StallRepository
	tickets *repository.TicketRepository
}

func NewTicketUsecase(stalls *repository.StallRepository, tickets *repository.TicketRepository) *TicketUsecase {
	return &TicketUsecase{stalls: stalls, tickets: tickets}
}

// Create は整理券を発行する。営業中（open）でない屋台では発行できない。
func (u *TicketUsecase) Create(ctx context.Context, stallID uint, comment *string) (*domain.Ticket, error) {
	stall, err := findStall(ctx, u.stalls, stallID)
	if err != nil {
		return nil, err
	}
	if stall.BusinessStatus != domain.BusinessStatusOpen {
		return nil, domain.StallNotOpen("営業中でない屋台では整理券を発行できません")
	}

	comment, err = normalizeComment(comment)
	if err != nil {
		return nil, err
	}

	token, err := newPublicToken()
	if err != nil {
		return nil, err
	}
	return u.tickets.Create(ctx, stallID, token, comment)
}

// normalizeComment は備考の前後の空白を取り、空なら null にする。長すぎるときはエラーにする。
func normalizeComment(comment *string) (*string, error) {
	if comment == nil {
		return nil, nil
	}
	c := strings.TrimSpace(*comment)
	if c == "" {
		return nil, nil
	}
	if utf8.RuneCountInString(c) > maxCommentLength {
		return nil, domain.InvalidInput(fmt.Sprintf("備考は %d 文字以内にしてください", maxCommentLength))
	}
	return &c, nil
}

// newPublicToken は推測されにくいランダムな文字列（128 ビット、URL に使える22文字）を作る。
func newPublicToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("public_token を作れません: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

// TicketDetail は来場者に見せる整理券。屋台の情報と、前に何人いるかを含む。
type TicketDetail struct {
	Ticket      *domain.Ticket
	Stall       *domain.Stall
	PeopleAhead int
}

// GetByPublicToken は public_token で整理券を取り、前に何人いるかを計算する。
func (u *TicketUsecase) GetByPublicToken(ctx context.Context, publicToken string) (*TicketDetail, error) {
	t, err := u.tickets.FindByPublicToken(ctx, publicToken)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.NotFound("整理券が見つかりません")
	}
	if err != nil {
		return nil, err
	}
	stall, err := u.stalls.FindByID(ctx, t.StallID)
	if err != nil {
		return nil, err
	}
	maxReceived, err := u.tickets.MaxReceivedOrderNumber(ctx, t.StallID)
	if err != nil {
		return nil, err
	}
	return &TicketDetail{Ticket: t, Stall: stall, PeopleAhead: peopleAhead(t, maxReceived)}, nil
}

// peopleAhead は前に何人いるかを返す。
//
//	前に何人 = 自分の order_number − 同じ屋台で received の最大 order_number − 1
//
// 受け取りに来なかった人は、後ろの番号が受け取り済みになった時点で数えなくなる。
// 後ろの番号の人が先に受け取るとマイナスになるので 0 にする。調理中（waiting）でなければ待っていないので 0。
func peopleAhead(t *domain.Ticket, maxReceivedOrderNumber int) int {
	if t.Status != domain.TicketStatusWaiting {
		return 0
	}
	return max(t.OrderNumber-maxReceivedOrderNumber-1, 0)
}

// 整理券一覧の件数（仕様書の limit と合わせる）。
const (
	defaultListLimit = 100
	maxListLimit     = 500
)

// ListInput は整理券一覧の絞り込みの条件。
type ListInput struct {
	Statuses []domain.TicketStatus
	Limit    *int   // 省略なら defaultListLimit
	Order    string // "asc"（省略時）か "desc"
}

// ListByStall は屋台の整理券を、status で絞り込んで order_number の順に返す。
// 整理券は10万件規模になるので、status の指定を必須にし、件数も制限する。
func (u *TicketUsecase) ListByStall(ctx context.Context, stallID uint, in ListInput) ([]domain.Ticket, error) {
	if len(in.Statuses) == 0 {
		return nil, domain.InvalidInput("status を1つ以上指定してください")
	}
	statuses := make([]domain.TicketStatus, 0, len(in.Statuses))
	seen := map[domain.TicketStatus]bool{}
	for _, s := range in.Statuses {
		if !s.Valid() {
			return nil, domain.InvalidInput("status は waiting / ready / received のどれかにしてください")
		}
		if !seen[s] {
			seen[s] = true
			statuses = append(statuses, s)
		}
	}

	limit := defaultListLimit
	if in.Limit != nil {
		limit = *in.Limit
	}
	if limit < 1 || limit > maxListLimit {
		return nil, domain.InvalidInput(fmt.Sprintf("limit は 1 から %d までにしてください", maxListLimit))
	}

	var descending bool
	switch in.Order {
	case "", "asc":
	case "desc":
		descending = true
	default:
		return nil, domain.InvalidInput("order は asc か desc にしてください")
	}

	if _, err := findStall(ctx, u.stalls, stallID); err != nil {
		return nil, err
	}
	return u.tickets.ListByStall(ctx, stallID, statuses, limit, descending)
}

// UpdateStatus は整理券の status を変える。押し間違いを戻せるように、どの status からどの status にも変えられる。
//
// called_at（調理完了にした時刻）は次のようにする。
//   - ready にしたとき：今の時刻を入れる
//   - waiting に戻したとき：空にする（まだ調理完了していないので）
//   - received にしたとき：そのまま残す（いつ呼び出したかの記録として）
//
// 今と同じ status を指定したときは、何も変えずにそのまま返す（ready を2回押しても呼び出し時刻は変わらない）。
func (u *TicketUsecase) UpdateStatus(ctx context.Context, id uint, status domain.TicketStatus) (*domain.Ticket, error) {
	if !status.Valid() {
		return nil, domain.InvalidInput("status は waiting / ready / received のどれかにしてください")
	}
	t, err := u.tickets.FindByID(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.NotFound("整理券が見つかりません")
	}
	if err != nil {
		return nil, err
	}
	if t.Status == status {
		return t, nil
	}

	calledAt := t.CalledAt
	switch status {
	case domain.TicketStatusReady:
		now := time.Now()
		calledAt = &now
	case domain.TicketStatusWaiting:
		calledAt = nil
	}
	return u.tickets.UpdateStatus(ctx, id, status, calledAt)
}
