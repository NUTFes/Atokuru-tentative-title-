package db

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

// stallCSVColumns は、屋台の CSV に必ずなければならない列。
// これ以外の列（purpose など、行の意図を書いたメモ）は読み飛ばす。
var stallCSVColumns = []string{"stall_number", "name", "hp_link", "business_status"}

// ReadStallsCSV は、屋台の CSV を読んで domain.Stall の一覧にする。
//
//   - 1行目はヘッダー。列の順番は問わない
//   - hp_link が空なら nil（NULL）にする
//   - password_hash は CSV からは読まない（nil のまま）
//   - business_status が決まっている値でない行、店番号が重複している行はエラーにする
func ReadStallsCSV(r io.Reader) ([]domain.Stall, error) {
	cr := csv.NewReader(r)

	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("CSV のヘッダーを読めません: %w", err)
	}
	// Excel で保存した CSV は先頭に BOM が付くので取り除く。
	if len(header) > 0 {
		header[0] = strings.TrimPrefix(header[0], "\ufeff")
	}
	col := map[string]int{}
	for i, name := range header {
		col[name] = i
	}
	for _, name := range stallCSVColumns {
		if _, ok := col[name]; !ok {
			return nil, fmt.Errorf("CSV に %s の列がありません（あるもの: %v）", name, header)
		}
	}

	var stalls []domain.Stall
	seen := map[string]int{}
	for line := 2; ; line++ {
		rec, err := cr.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("CSV の %d 行目を読めません: %w", line, err)
		}

		number := rec[col["stall_number"]]
		name := rec[col["name"]]
		status := domain.BusinessStatus(rec[col["business_status"]])
		switch {
		case number == "":
			return nil, fmt.Errorf("CSV の %d 行目: stall_number が空です", line)
		case name == "":
			return nil, fmt.Errorf("CSV の %d 行目（%s）: name が空です", line, number)
		case !status.Valid():
			return nil, fmt.Errorf("CSV の %d 行目（%s）: business_status %q は決まっている値ではありません", line, number, status)
		}
		if prev, ok := seen[number]; ok {
			return nil, fmt.Errorf("CSV の %d 行目: stall_number %s が %d 行目と重複しています", line, number, prev)
		}
		seen[number] = line

		var hpLink *string
		if v := rec[col["hp_link"]]; v != "" {
			hpLink = &v
		}
		stalls = append(stalls, domain.Stall{
			StallNumber:    number,
			Name:           name,
			HPLink:         hpLink,
			BusinessStatus: status,
		})
	}
	return stalls, nil
}

// SeedStalls は、屋台を DB に入れる。
//
// 店番号（stall_number）が同じ屋台がすでにあれば、店名・HP・営業状況を上書きする（upsert）。
// 消してから入れ直すと id が変わり、整理券の stall_id がずれるので、そうはしない。
// CSV にない屋台は消さずに残す。何度実行してもよい。
func SeedStalls(gdb *gorm.DB, stalls []domain.Stall) error {
	if len(stalls) == 0 {
		return errors.New("入れる屋台がありません")
	}
	err := gdb.Transaction(func(tx *gorm.DB) error {
		return tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "stall_number"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "hp_link", "business_status", "updated_at"}),
		}).CreateInBatches(&stalls, 100).Error
	})
	if err != nil {
		return fmt.Errorf("屋台を入れられません: %w", err)
	}
	return nil
}
