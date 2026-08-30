package ai

import (
	"strings"
	"testing"
)

func TestMarkdownToCSV(t *testing.T) {
	md := `| 序号 | 物品名称 | 金额(元) |
| :--- | :--- | :--- |
| 1 | CPU | 2,050.00 |
| 2 | 主板 | 1,099.00 |
`
	csv, err := markdownToCSV(md)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(csv, "物品名称") || !strings.Contains(csv, "CPU") {
		t.Fatalf("csv=%q", csv)
	}
	lines := strings.Split(strings.TrimSpace(csv), "\n")
	if len(lines) != 3 {
		t.Fatalf("lines=%d csv=%q", len(lines), csv)
	}
}
