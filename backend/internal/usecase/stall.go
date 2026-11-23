// Package usecase は残り人数の計算などの判断を書く。DB には repository を通してだけ触る。
package usecase

import (
	"context"
	"errors"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/repository"
)

// StallUsecase は屋台についての処理をまとめる。
type StallUsecase struct {
	stalls *repository.StallRepository
}

func NewStallUsecase(stalls *repository.StallRepository) *StallUsecase {
	return &StallUsecase{stalls: stalls}
}

// Get は ID で屋台を1件取る。
func (u *StallUsecase) Get(ctx context.Context, id uint) (*domain.Stall, error) {
	return findStall(ctx, u.stalls, id)
}

// findStall は屋台を取り、ないときは利用者に見せるメッセージ付きの not_found にする。
func findStall(ctx context.Context, stalls *repository.StallRepository, id uint) (*domain.Stall, error) {
	s, err := stalls.FindByID(ctx, id)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.NotFound("屋台が見つかりません")
	}
	if err != nil {
		return nil, err
	}
	return s, nil
}

// UpdateBusinessStatus は屋台の営業状況を変える。open / preparing / sold_out / closed のどれからどれにも変えられる。
func (u *StallUsecase) UpdateBusinessStatus(ctx context.Context, id uint, status domain.BusinessStatus) (*domain.Stall, error) {
	if !status.Valid() {
		return nil, domain.InvalidInput("business_status は open / preparing / sold_out / closed のどれかにしてください")
	}
	s, err := u.stalls.UpdateBusinessStatus(ctx, id, status)
	if errors.Is(err, domain.ErrNotFound) {
		return nil, domain.NotFound("屋台が見つかりません")
	}
	return s, err
}
