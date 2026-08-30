package ai

import "testing"

func TestParseMarkdownTable(t *testing.T) {
	md := `把如下表格内容添加在账本中：
| 序号 | 物品名称 | 规格型号 | 单位 | 数量 | 含税单价(元) | 金额(元) | 税率 | 税额(元) | 备注 |
| :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- | :--- |
| 1 | CPU | Intel Core i7-13700 散片 | 个 | 1 | 2,050.00 | 2,050.00 | 13% | 266.50 | 技术岗办公用 |
| 2 | 主板 | 华硕 TUF GAMING B760M-PLUS | 块 | 1 | 1,099.00 | 1,099.00 | 13% | 142.87 | 技术岗办公用 |
| 9 | 键鼠套装 | 罗技 MK245 Nano | 套 | 1 | 99.00 | 99.00 | 13% | 12.87 | 技术岗办公用 |
`
	headers, rows, err := parseMarkdownTable(md)
	if err != nil {
		t.Fatal(err)
	}
	if len(headers) != 10 {
		t.Fatalf("headers=%d want 10: %v", len(headers), headers)
	}
	if len(rows) != 3 {
		t.Fatalf("rows=%d want 3", len(rows))
	}
	if rows[0][1] != "CPU" || rows[2][1] != "键鼠套装" {
		t.Fatalf("unexpected cells: %v / %v", rows[0], rows[2])
	}
	plan := planFieldsFromHeaders(headers)
	if len(plan) != 9 {
		t.Fatalf("plan fields=%d want 9 (skip 序号)", len(plan))
	}
	maps := rowsToDataMaps(rows, plan)
	if maps[0]["item_name"] != "CPU" {
		t.Fatalf("item_name=%v", maps[0]["item_name"])
	}
	if maps[0]["unit_price"] != "2,050.00" {
		t.Fatalf("unit_price=%v", maps[0]["unit_price"])
	}
}
