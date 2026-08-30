package ai

import (
	"strings"
	"testing"
)

const sampleProcurementMD = `把如下表格内容添加在账本中：
| 序号 | 物品名称 | 规格型号 | 单位 | 数量 | 含税单价(元) | 金额(元) | 税率 | 税额(元) | 备注 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | CPU | Intel Core i7-13700 散片 | 个 | 1 | 2,050.00 | 2,050.00 | 13% | 266.50 | 技术岗办公用 |
| 2 | 主板 | 华硕 TUF GAMING B760M-PLUS | 块 | 1 | 1,099.00 | 1,099.00 | 13% | 142.87 | 技术岗办公用 |
| 3 | 内存 | 金士顿 FURY Beast 32GB DDR4 | 条 | 2 | 249.50 | 499.00 | 13% | 64.87 | 技术岗办公用 |
| 4 | 固态硬盘 | 致态 TiPlus7100 1TB | 个 | 1 | 569.00 | 569.00 | 13% | 73.97 | 技术岗办公用 |
| 5 | 散热器 | 利民 AX120 R SE | 个 | 1 | 89.00 | 89.00 | 13% | 11.57 | 技术岗办公用 |
| 6 | 电源 | 振华 冰山金蝶 HX530 | 个 | 1 | 399.00 | 399.00 | 13% | 51.87 | 技术岗办公用 |
| 7 | 机箱 | 先马 易大师 神光版 | 个 | 1 | 239.00 | 239.00 | 13% | 31.07 | 技术岗办公用 |
| 8 | 显示器 | AOC 24B35H | 台 | 1 | 499.00 | 499.00 | 13% | 64.87 | 技术岗办公用 |
| 9 | 键鼠套装 | 罗技 MK245 Nano | 套 | 1 | 99.00 | 99.00 | 13% | 12.87 | 技术岗办公用 |
`

func TestParseSampleProcurementTable_9rows(t *testing.T) {
	headers, rows, err := parseMarkdownTable(sampleProcurementMD)
	if err != nil {
		t.Fatal(err)
	}
	if len(headers) != 10 {
		t.Fatalf("headers=%d %v", len(headers), headers)
	}
	if len(rows) != 9 {
		t.Fatalf("rows=%d want 9", len(rows))
	}
	csv, err := markdownToCSV(sampleProcurementMD)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(csv), "\n")
	if len(lines) != 10 { // header + 9
		t.Fatalf("csv lines=%d want 10\n%s", len(lines), csv)
	}
}

func TestPreferRicherMarkdownFromSource(t *testing.T) {
	// Model often only re-emits the last row in tool args
	truncated := `| 序号 | 物品名称 | 规格型号 | 单位 | 数量 | 含税单价(元) | 金额(元) | 税率 | 税额(元) | 备注 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 9 | 键鼠套装 | 罗技 MK245 Nano | 套 | 1 | 99.00 | 99.00 | 13% | 12.87 | 技术岗办公用 |
`
	got := preferRicherMarkdown(truncated, sampleProcurementMD)
	_, rows, err := parseMarkdownTable(got)
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 9 {
		t.Fatalf("prefer richer failed: rows=%d", len(rows))
	}
}
