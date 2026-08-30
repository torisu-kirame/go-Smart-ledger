package auditexport

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/accounting"
	"github.com/smart-ledger/go-smart-ledger/go-backend/pkg/domain"
)

// TableRows is one sheet's entry lines for audit export (F48 + F49).
type TableRows struct {
	TableID   string
	TableName string
	Schema    domain.EntrySchema
	Rows      []EntryRow
}

// EntryRow is one EntryAdded event rendered for export.
type EntryRow struct {
	Seq      uint64
	SignerID string
	Data     map[string]string
	Hash     string
	At       time.Time
}

// Bundle is the full audit dataset before rendering.
type Bundle struct {
	LedgerID          string
	LedgerName        string
	BookkeepingMode   string
	MultiTableEnabled bool
	LatestSeq         uint64
	LatestRoot        string
	AnchorStatus      string
	IntegrityValid    bool
	ExportedAt        time.Time
	ExportedBy        string
	Tables            []TableRows
	Attachments       []accounting.Attachment
	ExternalAnchorTx  string
}

// Manifest is machine-readable audit metadata (JSON in zip).
type Manifest struct {
	LedgerID          string    `json:"ledgerId"`
	LedgerName        string    `json:"ledgerName"`
	ExportedAt        time.Time `json:"exportedAt"`
	ExportedBy        string    `json:"exportedBy"`
	BookkeepingMode   string    `json:"bookkeepingMode"`
	MultiTableEnabled bool      `json:"multiTableEnabled"`
	LatestSeq         uint64    `json:"latestSeq"`
	LatestRoot        string    `json:"latestRoot"`
	AnchorStatus      string    `json:"anchorStatus"`
	IntegrityValid    bool      `json:"integrityValid"`
	TableCount        int       `json:"tableCount"`
	EntryCount        int       `json:"entryCount"`
	AttachmentCount   int       `json:"attachmentCount"`
}

func (b *Bundle) Manifest() Manifest {
	entries := 0
	for _, t := range b.Tables {
		entries += len(t.Rows)
	}
	return Manifest{
		LedgerID:          b.LedgerID,
		LedgerName:        b.LedgerName,
		ExportedAt:        b.ExportedAt,
		ExportedBy:        b.ExportedBy,
		BookkeepingMode:   b.BookkeepingMode,
		MultiTableEnabled: b.MultiTableEnabled,
		LatestSeq:         b.LatestSeq,
		LatestRoot:        b.LatestRoot,
		AnchorStatus:      b.AnchorStatus,
		IntegrityValid:    b.IntegrityValid,
		TableCount:        len(b.Tables),
		EntryCount:        entries,
		AttachmentCount:   len(b.Attachments),
	}
}

// BuildExcel returns multi-sheet xlsx bytes.
func BuildExcel(b *Bundle) ([]byte, error) {
	return buildWorkbook(b)
}

// BuildPDF returns a one-page audit summary PDF.
func BuildPDF(b *Bundle) ([]byte, error) {
	return buildSummaryPDF(b)
}

// BuildZip returns xlsx + pdf + manifest.json.
func BuildZip(b *Bundle) ([]byte, error) {
	xlsx, err := BuildExcel(b)
	if err != nil {
		return nil, err
	}
	pdf, err := BuildPDF(b)
	if err != nil {
		return nil, err
	}
	manifest, _ := json.MarshalIndent(b.Manifest(), "", "  ")
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	stamp := b.ExportedAt.Format("20060102-150405")
	add := func(name string, data []byte) error {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write(data)
		return err
	}
	if err := add(fmt.Sprintf("audit-%s-manifest.json", stamp), manifest); err != nil {
		return nil, err
	}
	if err := add(fmt.Sprintf("audit-%s.xlsx", stamp), xlsx); err != nil {
		return nil, err
	}
	if err := add(fmt.Sprintf("audit-%s-summary.pdf", stamp), pdf); err != nil {
		return nil, err
	}
	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
