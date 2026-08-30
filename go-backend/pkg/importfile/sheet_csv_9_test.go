package importfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPrepareCSVForNewSheet_9rows(t *testing.T) {
	csv := "序号,物品名称,规格型号,单位,数量,含税单价(元),金额(元),税率,税额(元),备注\n" +
		"1,CPU,Intel Core i7-13700 散片,个,1,\"2,050.00\",\"2,050.00\",13%,266.50,技术岗办公用\n" +
		"2,主板,华硕 TUF GAMING B760M-PLUS,块,1,\"1,099.00\",\"1,099.00\",13%,142.87,技术岗办公用\n" +
		"3,内存,金士顿 FURY Beast 32GB DDR4,条,2,249.50,499.00,13%,64.87,技术岗办公用\n" +
		"4,固态硬盘,致态 TiPlus7100 1TB,个,1,569.00,569.00,13%,73.97,技术岗办公用\n" +
		"5,散热器,利民 AX120 R SE,个,1,89.00,89.00,13%,11.57,技术岗办公用\n" +
		"6,电源,振华 冰山金蝶 HX530,个,1,399.00,399.00,13%,51.87,技术岗办公用\n" +
		"7,机箱,先马 易大师 神光版,个,1,239.00,239.00,13%,31.07,技术岗办公用\n" +
		"8,显示器,AOC 24B35H,台,1,499.00,499.00,13%,64.87,技术岗办公用\n" +
		"9,键鼠套装,罗技 MK245 Nano,套,1,99.00,99.00,13%,12.87,技术岗办公用\n"
	schema, rows, err := PrepareCSVForNewSheet([]byte(csv))
	if err != nil {
		t.Fatal(err)
	}
	if len(schema.Fields) != 9 {
		t.Fatalf("fields=%d want 9", len(schema.Fields))
	}
	if len(rows) != 9 {
		t.Fatalf("rows=%d want 9", len(rows))
	}
	valid := 0
	for _, r := range rows {
		if r.Error == "" {
			valid++
		} else {
			t.Logf("row %d err: %s cells=%v", r.Line, r.Error, r.Cells)
		}
	}
	if valid != 9 {
		t.Fatalf("valid=%d want 9", valid)
	}
	_ = filepath.Join
	_ = os.TempDir
}
