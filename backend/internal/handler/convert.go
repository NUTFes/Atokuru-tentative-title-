package handler

import (
	"math"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/gen"
)

// domain の entity を、仕様書の形（gen の型）に詰め替える。

func toStall(s *domain.Stall) gen.Stall {
	return gen.Stall{
		Id:             int64(s.ID),
		Name:           s.Name,
		StallNumber:    s.StallNumber,
		HpLink:         s.HPLink,
		BusinessStatus: gen.BusinessStatus(s.BusinessStatus),
	}
}

// toID はパスの ID を uint にする。0 以下や大きすぎる値は、存在しない ID として 0 を返す（0 の ID はない）。
func toID(id int64) uint {
	if id <= 0 || uint64(id) > math.MaxUint32 {
		return 0
	}
	return uint(id)
}
