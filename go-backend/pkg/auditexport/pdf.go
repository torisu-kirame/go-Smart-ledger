package auditexport

import (
	"bytes"
	"fmt"
)

// buildSummaryPDF emits a minimal PDF 1.4 one-page text summary (no external fonts).
func buildSummaryPDF(b *Bundle) ([]byte, error) {
	lines := []string{
		"Smart Ledger Audit Summary",
		fmt.Sprintf("Ledger: %s (%s)", b.LedgerName, b.LedgerID),
		fmt.Sprintf("Exported: %s by %s", b.ExportedAt.Format("2006-01-02 15:04:05 UTC"), b.ExportedBy),
		fmt.Sprintf("Mode: %s | Multi-table: %v", b.BookkeepingMode, b.MultiTableEnabled),
		fmt.Sprintf("Latest seq: %d | Merkle root: %s", b.LatestSeq, truncateRoot(b.LatestRoot)),
		fmt.Sprintf("Anchor: %s | Integrity valid: %v", b.AnchorStatus, b.IntegrityValid),
		fmt.Sprintf("Tables: %d | Entries: %d | Attachments: %d",
			len(b.Tables), entryTotal(b), len(b.Attachments)),
	}
	if b.ExternalAnchorTx != "" {
		lines = append(lines, "External anchor: "+b.ExternalAnchorTx)
	}
	lines = append(lines, "", "This package is for compliance review. See Excel workbook for full detail.")
	return minimalTextPDF(lines), nil
}

func truncateRoot(s string) string {
	if len(s) <= 48 {
		return s
	}
	return s[:24] + "..." + s[len(s)-12:]
}

func minimalTextPDF(lines []string) []byte {
	var content bytes.Buffer
	content.WriteString("BT\n/F1 11 Tf\n50 780 Td\n")
	for i, line := range lines {
		if i > 0 {
			content.WriteString("0 -16 Td\n")
		}
		content.WriteString(fmt.Sprintf("(%s) Tj\n", pdfEscape(line)))
	}
	content.WriteString("ET\n")
	contentStr := content.String()

	var buf bytes.Buffer
	buf.WriteString("%PDF-1.4\n")
	offsets := []int{0}
	writeObj := func(n int, body string) {
		offsets = append(offsets, buf.Len())
		fmt.Fprintf(&buf, "%d 0 obj\n%s\nendobj\n", n, body)
	}
	writeObj(1, "<< /Type /Catalog /Pages 2 0 R >>")
	writeObj(2, "<< /Type /Pages /Kids [3 0 R] /Count 1 >>")
	writeObj(3, "<< /Type /Page /Parent 2 0 R /MediaBox [0 0 595 842] /Contents 4 0 R /Resources << /Font << /F1 5 0 R >> >> >>")
	writeObj(4, fmt.Sprintf("<< /Length %d >>\nstream\n%s\nendstream", len(contentStr), contentStr))
	writeObj(5, "<< /Type /Font /Subtype /Type1 /BaseFont /Helvetica >>")

	xrefPos := buf.Len()
	fmt.Fprintf(&buf, "xref\n0 %d\n", len(offsets))
	buf.WriteString("0000000000 65535 f \n")
	for i := 1; i < len(offsets); i++ {
		fmt.Fprintf(&buf, "%010d 00000 n \n", offsets[i])
	}
	fmt.Fprintf(&buf, "trailer\n<< /Size %d /Root 1 0 R >>\nstartxref\n%d\n%%EOF\n", len(offsets), xrefPos)
	return buf.Bytes()
}

func pdfEscape(s string) string {
	var b bytes.Buffer
	for _, r := range s {
		switch r {
		case '(', ')', '\\':
			b.WriteByte('\\')
			b.WriteRune(r)
		default:
			if r < 32 || r > 126 {
				continue
			}
			b.WriteRune(r)
		}
	}
	return b.String()
}
