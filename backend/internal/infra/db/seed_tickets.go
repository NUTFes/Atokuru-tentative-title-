package db

import (
	"encoding/base64"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

// ticketBatchSize は、整理券を1回の INSERT で入れる件数。
// 列が8つなので、1回の SQL の変数は 4,000 個になり、SQLite の上限（32,766）より十分少ない。
const ticketBatchSize = 500

// openWaitingPattern は、営業中（open）の屋台に割り当てる待ち人数（waiting の件数）。
// 混雑度の緑（0〜4人）・黄（5〜9人）・赤（10人以上）と、大行列（30人以上）が全部出るようにする。
var openWaitingPattern = []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 12, 15, 20, 30, 45}

// ticketComments は、整理券の備考に入れる値。画面の確認用の極端な値も混ぜる。
var ticketComments = []string{
	"大盛り",
	"ソース抜き",
	"2個",
	"辛さ控えめ・ネギ多め",
	"🍜×2",
	"<b>太字</b><script>alert(1)</script>",
	`"引用" & 'アンド'`,
	"1行目\n2行目",
	"ｱﾚﾙｷﾞｰ: ｴﾋﾞ",
	strings.Repeat("長", 200), // 備考の上限（200文字）ちょうど
}

// commentRate は、備考がある整理券の割合。
const commentRate = 0.2

// festivalStarts は、学園祭の各日の受付開始時刻（日本時間）。各日 festivalHours 時間受け付ける。
var festivalStarts = []time.Time{
	time.Date(2026, 11, 21, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
	time.Date(2026, 11, 22, 10, 0, 0, 0, time.FixedZone("JST", 9*60*60)),
}

const festivalHours = 6

// stallPlan は、1つの屋台に入れる整理券の内訳。
type stallPlan struct {
	total   int // 整理券の数
	waiting int // 最後の waiting の件数
	ready   int // waiting の直前の ready の件数
	stale   int // received の中に混ぜる、取りに来ていない ready の件数
}

// minTotal は、内訳を満たすのに最低限必要な件数。最後の received（前に何人の基準）を必ず1枚残す。
func (p stallPlan) minTotal() int {
	return p.waiting + p.ready + p.stale + 1
}

// GenerateTickets は、stalls（DB に入っていて ID があるもの）に total 件の整理券を割り振って作る。
// DB には触らない。seed が同じなら、同じ整理券（public_token を含む）を作る。
//
// 屋台ごとの整理券は、番号の古い方から received → ready → waiting の順に並べる。
// こうすると「前に何人 = 自分の番号 − received の最大番号 − 1」が意味のある値になる。
// 一部の屋台には、received の間に取りに来ていない古い ready を混ぜる。
//
// 営業状況ごとの待ち人数：
//   - open：openWaitingPattern の値（混雑度の色が全部出る）
//   - sold_out：0〜3人（売り切れたが、すでに受けた注文は作っている）
//   - preparing / closed：0人
func GenerateTickets(stalls []domain.Stall, total int, seed uint64) ([]domain.Ticket, error) {
	if len(stalls) == 0 {
		return nil, errors.New("整理券を割り振る屋台がありません")
	}
	for _, s := range stalls {
		if s.ID == 0 {
			return nil, fmt.Errorf("屋台 %s に ID がありません（DB から読んだ屋台を渡してください）", s.StallNumber)
		}
	}

	// 渡された順番によらず同じデータになるように、店番号の順に並べる。
	stalls = slices.Clone(stalls)
	slices.SortFunc(stalls, func(a, b domain.Stall) int { return strings.Compare(a.StallNumber, b.StallNumber) })

	var key [32]byte
	binary.LittleEndian.PutUint64(key[:], seed)
	src := rand.NewChaCha8(key)
	rng := rand.New(src)

	plans, err := planStalls(stalls, total, rng)
	if err != nil {
		return nil, err
	}

	tickets := make([]domain.Ticket, 0, total)
	tokenBytes := make([]byte, 16)
	for i, s := range stalls {
		p := plans[i]
		statuses := ticketStatuses(p, rng)
		created := createdTimes(p.total, rng)
		for j := range p.total {
			// usecase の発券と同じ形（16バイトを base64url にした22文字）にする。
			_, _ = src.Read(tokenBytes)
			t := domain.Ticket{
				PublicToken: base64.RawURLEncoding.EncodeToString(tokenBytes),
				StallID:     s.ID,
				OrderNumber: j + 1,
				Status:      statuses[j],
				CreatedAt:   created[j],
				UpdatedAt:   created[j],
			}
			if rng.Float64() < commentRate {
				c := ticketComments[rng.IntN(len(ticketComments))]
				t.Comment = &c
			}
			if t.Status != domain.TicketStatusWaiting {
				called := clampToEnd(created[j].Add(randMinutes(rng, 3, 15)))
				t.CalledAt = &called
				t.UpdatedAt = called
				if t.Status == domain.TicketStatusReceived {
					t.UpdatedAt = clampToEnd(called.Add(randMinutes(rng, 1, 10)))
				}
			}
			tickets = append(tickets, t)
		}
	}
	return tickets, nil
}

// planStalls は、屋台ごとの内訳を決める。
// まず待ち人数などの最低限の件数を割り当て、残りを屋台の人気（重み）に応じて割り振る。
func planStalls(stalls []domain.Stall, total int, rng *rand.Rand) ([]stallPlan, error) {
	plans := make([]stallPlan, len(stalls))

	// open の屋台の並びを混ぜてから、待ち人数のパターンを順に割り当てる。
	// （店番号の順のままだと、A-01 などの極端な店名の店がいつも同じ待ち人数になる）
	var open []int
	for i, s := range stalls {
		if s.BusinessStatus == domain.BusinessStatusOpen {
			open = append(open, i)
		}
	}
	for k, idx := range rng.Perm(len(open)) {
		plans[open[idx]].waiting = openWaitingPattern[k%len(openWaitingPattern)]
	}

	minSum := 0
	for i, s := range stalls {
		p := &plans[i]
		switch s.BusinessStatus {
		case domain.BusinessStatusOpen:
			p.ready = rng.IntN(4)
		case domain.BusinessStatusSoldOut:
			p.waiting = rng.IntN(4)
			p.ready = rng.IntN(3)
		}
		// 4店に1店くらいは、取りに来ていない古い ready を1〜2枚混ぜる。
		if rng.IntN(4) == 0 {
			p.stale = 1 + rng.IntN(2)
		}
		minSum += p.minTotal()
	}
	if total < minSum {
		return nil, fmt.Errorf("整理券が %d 件では足りません（%d 店の待ち人数などを満たすには %d 件以上必要です）", total, len(stalls), minSum)
	}

	// 人気の偏り。対数正規分布にすると、少数の人気店に件数が集まる。
	weights := make([]float64, len(stalls))
	var weightSum float64
	for i := range weights {
		weights[i] = math.Exp(rng.NormFloat64() * 0.7)
		weightSum += weights[i]
	}
	rest := total - minSum
	assigned := 0
	for i := range plans {
		extra := int(float64(rest) * weights[i] / weightSum)
		plans[i].total = plans[i].minTotal() + extra
		assigned += extra
	}
	// 切り捨てた端数（屋台の数より少ない）を、先頭の屋台から1件ずつ足す。
	for i := 0; assigned < rest; i++ {
		plans[i%len(plans)].total++
		assigned++
	}
	return plans, nil
}

// ticketStatuses は、番号順（添字0が1番）の status を作る。
func ticketStatuses(p stallPlan, rng *rand.Rand) []domain.TicketStatus {
	statuses := make([]domain.TicketStatus, p.total)
	received := p.total - p.waiting - p.ready
	for j := range statuses {
		switch {
		case j < received:
			statuses[j] = domain.TicketStatusReceived
		case j < received+p.ready:
			statuses[j] = domain.TicketStatusReady
		default:
			statuses[j] = domain.TicketStatusWaiting
		}
	}
	// 古い ready は、最後の received より前に置く（received の最大番号は変えない）。
	for _, j := range rng.Perm(received - 1)[:p.stale] {
		statuses[j] = domain.TicketStatusReady
	}
	return statuses
}

// createdTimes は、n 枚の発券時刻を、学園祭の受付時間の中に古い順で作る。
func createdTimes(n int, rng *rand.Rand) []time.Time {
	span := time.Duration(festivalHours) * time.Hour
	offsets := make([]time.Duration, n)
	for j := range offsets {
		offsets[j] = time.Duration(rng.Int64N(int64(span) * int64(len(festivalStarts))))
	}
	slices.Sort(offsets)
	times := make([]time.Time, n)
	for j, off := range offsets {
		day := int(off / span)
		times[j] = festivalStarts[day].Add(off % span).Truncate(time.Second)
	}
	return times
}

func randMinutes(rng *rand.Rand, min, max int) time.Duration {
	return time.Duration(min+rng.IntN(max-min+1))*time.Minute + time.Duration(rng.IntN(60))*time.Second
}

// clampToEnd は、時刻が最終日の受付終了より後にならないようにする。
func clampToEnd(t time.Time) time.Time {
	end := festivalStarts[len(festivalStarts)-1].Add(festivalHours * time.Hour)
	if t.After(end) {
		return end
	}
	return t
}

// ListStalls は、DB にある屋台を店番号の順に返す。整理券を割り振る相手を決めるのに使う。
func ListStalls(gdb *gorm.DB) ([]domain.Stall, error) {
	var stalls []domain.Stall
	if err := gdb.Order("stall_number").Find(&stalls).Error; err != nil {
		return nil, fmt.Errorf("屋台を読めません: %w", err)
	}
	return stalls, nil
}

// SeedTickets は、整理券を DB に入れる。
//
// すでに整理券があるときは、reset が false ならエラーにする。true なら全部消してから入れる。
// Turso はトランザクションが長いと切れるので、全体を1つのトランザクションにはせず、
// ticketBatchSize 件ずつの INSERT で入れる。途中で失敗したら、reset を付けて入れ直す。
// progress は、何件入れたかを ticketBatchSize 件ごとに受け取る（nil でよい）。
func SeedTickets(gdb *gorm.DB, tickets []domain.Ticket, reset bool, progress func(done int)) error {
	if len(tickets) == 0 {
		return errors.New("入れる整理券がありません")
	}

	var existing int64
	if err := gdb.Model(&domain.Ticket{}).Count(&existing).Error; err != nil {
		return fmt.Errorf("整理券の件数を数えられません: %w", err)
	}
	if existing > 0 {
		if !reset {
			return fmt.Errorf("整理券がすでに %d 件あります（消してから入れ直すときは -reset を付けてください）", existing)
		}
		if err := gdb.Where("1 = 1").Delete(&domain.Ticket{}).Error; err != nil {
			return fmt.Errorf("整理券を消せません: %w", err)
		}
	}

	// GORM は Create を既定でトランザクションに包むので、外す。屋台（Stall）は入れ直さない。
	tx := gdb.Session(&gorm.Session{SkipDefaultTransaction: true}).Omit(clause.Associations)
	for start := 0; start < len(tickets); start += ticketBatchSize {
		end := min(start+ticketBatchSize, len(tickets))
		if err := tx.Create(tickets[start:end]).Error; err != nil {
			return fmt.Errorf("整理券の %d〜%d 件目を入れられません: %w", start+1, end, err)
		}
		if progress != nil {
			progress(end)
		}
	}
	return nil
}

// StallWaiting は、屋台の営業状況と、waiting の整理券の数。
type StallWaiting struct {
	StallNumber    string
	BusinessStatus domain.BusinessStatus
	Waiting        int
}

// TicketSummary は、入れた整理券を確かめるための集計。中身を全部読まずに済むようにする。
type TicketSummary struct {
	ByStatus map[domain.TicketStatus]int64
	Stalls   []StallWaiting
}

// SummarizeTickets は、status ごとの件数と、屋台ごとの待ち人数を数える。
func SummarizeTickets(gdb *gorm.DB) (*TicketSummary, error) {
	var rows []struct {
		Status domain.TicketStatus
		N      int64
	}
	if err := gdb.Model(&domain.Ticket{}).Select("status, COUNT(*) AS n").Group("status").Scan(&rows).Error; err != nil {
		return nil, fmt.Errorf("整理券を数えられません: %w", err)
	}
	sum := &TicketSummary{ByStatus: map[domain.TicketStatus]int64{}}
	for _, r := range rows {
		sum.ByStatus[r.Status] = r.N
	}

	err := gdb.Table("stalls AS s").
		Select("s.stall_number, s.business_status, COUNT(t.id) AS waiting").
		Joins("LEFT JOIN tickets AS t ON t.stall_id = s.id AND t.status = ?", domain.TicketStatusWaiting).
		Group("s.id").
		Order("s.stall_number").
		Scan(&sum.Stalls).Error
	if err != nil {
		return nil, fmt.Errorf("屋台ごとの待ち人数を数えられません: %w", err)
	}
	return sum, nil
}
