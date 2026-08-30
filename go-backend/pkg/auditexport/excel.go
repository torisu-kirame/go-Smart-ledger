package auditexport

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/xuri/excelize/v2"
)

var invalidSheet = regexp.MustCompile(`[\[\]:*?/\\]`)

func buildWorkbook(b *Bundle) ([]byte, error) {
	f := excelize.NewFile()
	defer f.Close()

	if err := writeCoverSheet(f, b); err != nil {
		return nil, err
	}
	if err := writeAttachmentsSheet(f, b); err != nil {
		return nil, err
	}
	for _, tbl := range b.Tables {
		if err := writeTableSheet(f, b, tbl); err != nil {
			return nil, err
		}
	}

	if idx, _ := f.GetSheetIndex("Sheet1"); idx > 0 {
		if _, err := f.GetSheetIndex("审计封面"); err == nil {
			_ = f.DeleteSheet("Sheet1")
		}
	}
	return writeToBuffer(f)
}

func writeToBuffer(f *excelize.File) ([]byte, error) {
	from, err := f.WriteToBuffer()
	if err != nil {
		return nil, err
	}
	return from.Bytes(), nil
}

func writeCoverSheet(f *excelize.File, b *Bundle) error {
	sheet := "审计封面"
	idx, err := f.NewSheet(sheet)
	if err != nil {
		return err
	}
	f.SetActiveSheet(idx)
	rows := [][]string{
		{"Smart Ledger 审计包", ""},
		{"账本 ID", b.LedgerID},
		{"账本名称", b.LedgerName},
		{"导出时间", b.ExportedAt.Format("2006-01-02 15:04:05 UTC")},
		{"导出人", b.ExportedBy},
		{"记账模式", b.BookkeepingMode},
		{"多表模式", fmt.Sprintf("%v", b.MultiTableEnabled)},
		{"最新序号", fmt.Sprintf("%d", b.LatestSeq)},
		{"Merkle 根", b.LatestRoot},
		{"锚定状态", b.AnchorStatus},
		{"完整性校验", fmt.Sprintf("%v", b.IntegrityValid)},
		{"子表数量", fmt.Sprintf("%d", len(b.Tables))},
		{"流水笔数", fmt.Sprintf("%d", entryTotal(b))},
		{"附件数量", fmt.Sprintf("%d", len(b.Attachments))},
	}
	if b.ExternalAnchorTx != "" {
		rows = append(rows, []string{"链外锚定 Tx", b.ExternalAnchorTx})
	}
	for i, row := range rows {
		for j, v := range row {
			cell, _ := excelize.CoordinatesToCellName(j+1, i+1)
			_ = f.SetCellValue(sheet, cell, v)
		}
	}
	_ = f.DeleteSheet("Sheet1")
	return nil
}

func entryTotal(b *Bundle) int {
	n := 0
	for _, t := range b.Tables {
		n += len(t.Rows)
	}
	return n
}

func writeAttachmentsSheet(f *excelize.File, b *Bundle) error {
	sheet := "凭证附件"
	_, _ = f.NewSheet(sheet)
	headers := []string{
		"附件ID", "表ID", "流水Seq", "文件名", "CID", "部门", "项目", "往来",
		"大小", "上传人", "上传时间",
	}
	writeHeaderRow(f, sheet, headers)
	for i, a := range b.Attachments {
		dept, proj, cp := "", "", ""
		if a.Auxiliary != nil {
			dept, proj, cp = a.Auxiliary.Department, a.Auxiliary.Project, a.Auxiliary.Counterparty
		}
		row := []any{
			a.ID, a.TableID, a.EntrySeq, a.Filename, a.CID,
			dept, proj, cp, a.Size, a.UploadedBy, a.CreatedAt.Format("2006-01-02 15:04:05"),
		}
		writeDataRow(f, sheet, i+2, row)
	}
	return nil
}

func writeTableSheet(f *excelize.File, b *Bundle, tbl TableRows) error {
	name := sanitizeSheetName(tbl.TableName)
	if name == "" {
		name = tbl.TableID
	}
	// avoid duplicate sheet names
	base := name
	for n := 1; ; n++ {
		if idx, err := f.GetSheetIndex(name); err != nil || idx == 0 {
			break
		}
		name = fmt.Sprintf("%s_%d", truncate(base, 28), n)
	}
	_, _ = f.NewSheet(name)
	headers := []string{"Seq", "签名者", "事件哈希", "记账时间"}
	for _, fd := range tbl.Schema.Fields {
		headers = append(headers, fd.Label)
	}
	writeHeaderRow(f, name, headers)
	for i, row := range tbl.Rows {
		vals := []any{row.Seq, row.SignerID, row.Hash, row.At.Format("2006-01-02 15:04:05")}
		for _, fd := range tbl.Schema.Fields {
			vals = append(vals, row.Data[fd.Key])
		}
		writeDataRow(f, name, i+2, vals)
	}
	return nil
}

func writeHeaderRow(f *excelize.File, sheet string, headers []string) {
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
	}
}

func writeDataRow(f *excelize.File, sheet string, row int, vals []any) {
	for i, v := range vals {
		cell, _ := excelize.CoordinatesToCellName(i+1, row)
		_ = f.SetCellValue(sheet, cell, v)
	}
}

func sanitizeSheetName(s string) string {
	s = strings.TrimSpace(invalidSheet.ReplaceAllString(s, "_"))
	if s == "" {
		return "表"
	}
	return truncate(s, 31)
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}

