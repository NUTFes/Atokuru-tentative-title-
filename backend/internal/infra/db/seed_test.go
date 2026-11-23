package db

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gorm.io/gorm"

	"github.com/NUTFes/Atokuru-tentative-title-/backend/internal/domain"
)

func openMigrated(t *testing.T) *gorm.DB {
	t.Helper()
	gdb, err := Open("file:"+filepath.Join(t.TempDir(), "test.db"), "")
	if err != nil {
		t.Fatal(err)
	}
	if err := Migrate(gdb); err != nil {
		t.Fatal(err)
	}
	return gdb
}

func countStalls(t *testing.T, gdb *gorm.DB) int64 {
	t.Helper()
	var n int64
	if err := gdb.Model(&domain.Stall{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

// testdata/stalls.csv が読めて、80店が決まった配分で入るかを確かめる。2回入れても80店のまま。
func TestSeedStalls_TestdataCSV(t *testing.T) {
	f, err := os.Open(filepath.Join("..", "..", "..", "testdata", "stalls.csv"))
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	stalls, err := ReadStallsCSV(f)
	if err != nil {
		t.Fatal(err)
	}

	gdb := openMigrated(t)
	for i := 0; i < 2; i++ {
		if err := SeedStalls(gdb, stalls); err != nil {
			t.Fatalf("%d 回目: %v", i+1, err)
		}
		if n := countStalls(t, gdb); n != 80 {
			t.Fatalf("%d 回目: 屋台が %d 店です（80店のはず）", i+1, n)
		}
	}

	var rows []struct {
		BusinessStatus string
		N              int
	}
	if err := gdb.Model(&domain.Stall{}).Select("business_status, COUNT(*) AS n").Group("business_status").Scan(&rows).Error; err != nil {
		t.Fatal(err)
	}
	got := map[string]int{}
	for _, r := range rows {
		got[r.BusinessStatus] = r.N
	}
	want := map[string]int{"open": 50, "preparing": 10, "sold_out": 10, "closed": 10}
	for k, v := range want {
		if got[k] != v {
			t.Errorf("business_status %s が %d 店です（%d 店のはず）", k, got[k], v)
		}
	}

	var nullHP int64
	if err := gdb.Model(&domain.Stall{}).Where("hp_link IS NULL").Count(&nullHP).Error; err != nil {
		t.Fatal(err)
	}
	if nullHP != 20 {
		t.Errorf("hp_link が NULL の屋台が %d 店です（20店のはず）", nullHP)
	}

	// " とカンマを含む店名が、CSV の引用を外した形で入っているか。
	var a04 domain.Stall
	if err := gdb.Where("stall_number = ?", "A-04").First(&a04).Error; err != nil {
		t.Fatal(err)
	}
	if want := `Café "Mom's" & Co., Ltd.`; a04.Name != want || a04.HPLink != nil {
		t.Errorf("A-04: name %q, hp_link %v（name %q、hp_link nil のはず）", a04.Name, a04.HPLink, want)
	}
}

// 2回目に入れたときは id を変えずに上書きし、CSV にない屋台は残すかを確かめる。
func TestSeedStalls_Upsert(t *testing.T) {
	gdb := openMigrated(t)
	first, err := ReadStallsCSV(strings.NewReader("stall_number,name,hp_link,business_status\nA-01,旧店名,https://example.com,open\nA-02,残る店,,closed\n"))
	if err != nil {
		t.Fatal(err)
	}
	if err := SeedStalls(gdb, first); err != nil {
		t.Fatal(err)
	}
	var before domain.Stall
	if err := gdb.Where("stall_number = ?", "A-01").First(&before).Error; err != nil {
		t.Fatal(err)
	}

	second, err := ReadStallsCSV(strings.NewReader("stall_number,name,hp_link,business_status\nA-01,新店名,,sold_out\n"))
	if err != nil {
		t.Fatal(err)
	}
	if err := SeedStalls(gdb, second); err != nil {
		t.Fatal(err)
	}
	var after domain.Stall
	if err := gdb.Where("stall_number = ?", "A-01").First(&after).Error; err != nil {
		t.Fatal(err)
	}
	if after.ID != before.ID {
		t.Errorf("id が %d から %d に変わりました", before.ID, after.ID)
	}
	if after.Name != "新店名" || after.HPLink != nil || after.BusinessStatus != domain.BusinessStatusSoldOut {
		t.Errorf("上書きされていません: %+v", after)
	}
	if n := countStalls(t, gdb); n != 2 {
		t.Errorf("屋台が %d 店です（CSV にない A-02 も残って2店のはず）", n)
	}
}

func TestReadStallsCSV(t *testing.T) {
	t.Run("BOM と余分な列があっても読める", func(t *testing.T) {
		stalls, err := ReadStallsCSV(strings.NewReader("\ufeffpurpose,business_status,name,stall_number,hp_link\nメモ,open,店,A-01,\n"))
		if err != nil {
			t.Fatal(err)
		}
		if len(stalls) != 1 || stalls[0].StallNumber != "A-01" || stalls[0].HPLink != nil || stalls[0].PasswordHash != nil {
			t.Errorf("読んだ値が違います: %+v", stalls)
		}
	})

	errCases := map[string]string{
		"列が足りない":      "stall_number,name,business_status\nA-01,店,open\n",
		"決まっていない営業状況": "stall_number,name,hp_link,business_status\nA-01,店,,busy\n",
		"店番号の重複":      "stall_number,name,hp_link,business_status\nA-01,店,,open\nA-01,店2,,open\n",
		"店名が空":        "stall_number,name,hp_link,business_status\nA-01,,,open\n",
		"店番号が空":       "stall_number,name,hp_link,business_status\n,店,,open\n",
		"列の数が行によって違う": "stall_number,name,hp_link,business_status\nA-01,店,open\n",
	}
	for name, input := range errCases {
		t.Run(name, func(t *testing.T) {
			if _, err := ReadStallsCSV(strings.NewReader(input)); err == nil {
				t.Error("エラーになるはずです")
			}
		})
	}
}
