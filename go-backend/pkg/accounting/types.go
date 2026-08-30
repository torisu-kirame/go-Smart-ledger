package accounting

import "time"

// Event names for attachment lifecycle (stored on chain).
const (
	EventAttachmentLinked     = "AttachmentLinked"
	EventAttachmentAuxUpdated = "AttachmentAuxUpdated"
)

// AuxiliaryDims are optional analytic tags on attachments.
type AuxiliaryDims struct {
	Department   string `json:"department,omitempty"`   // 部门
	Project      string `json:"project,omitempty"`      // 项目
	Counterparty string `json:"counterparty,omitempty"` // 往来
}

// Attachment links a file (IPFS CID) to an entry event seq.
type Attachment struct {
	ID         string         `json:"id"`
	TableID    string         `json:"tableId,omitempty"`
	EntrySeq   uint64         `json:"entrySeq"`
	Filename   string         `json:"filename"`
	MimeType   string         `json:"mimeType,omitempty"`
	Size       int64          `json:"size,omitempty"`
	CID        string         `json:"cid"`
	Ref        string         `json:"ref,omitempty"`
	Auxiliary  *AuxiliaryDims `json:"auxiliary,omitempty"`
	UploadedBy string         `json:"uploadedBy"`
	CreatedAt  time.Time      `json:"createdAt"`
}
