package storage

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/smart-ledger/go-smart-ledger/backend/pkg/ipfsclient"
)

// PutResult is returned after dual-write backup.
type PutResult struct {
	Ref     string `json:"ref"`
	IPFSCID string `json:"ipfsCid,omitempty"`
}

// DualBackup writes encrypted blobs to disk and optionally IPFS.
type DualBackup struct {
	disk *DiskBackup
	ipfs *ipfsclient.Client
}

func NewDualBackup(disk *DiskBackup, ipfs *ipfsclient.Client) *DualBackup {
	return &DualBackup{disk: disk, ipfs: ipfs}
}

func (d *DualBackup) Put(ctx context.Context, ledgerID, password string, plain []byte) (*PutResult, error) {
	ref, err := d.disk.Put(ctx, ledgerID, password, plain)
	if err != nil {
		return nil, err
	}
	res := &PutResult{Ref: ref}
	if d.ipfs == nil || !d.ipfs.Enabled() {
		return res, nil
	}
	cipher, err := d.readCipher(ref)
	if err != nil {
		return res, nil
	}
	cid, err := d.ipfs.Add(ctx, cipher, true)
	if err != nil {
		return res, nil
	}
	_ = d.ipfs.Pin(ctx, cid)
	res.IPFSCID = cid
	_ = d.writeSidecar(ref, cid)
	return res, nil
}

func (d *DualBackup) Get(ctx context.Context, ref, password, ipfsCID string) ([]byte, error) {
	plain, err := d.disk.Get(ctx, ref, password)
	if err == nil {
		return plain, nil
	}
	if d.ipfs == nil || !d.ipfs.Enabled() {
		return nil, err
	}
	cid := ipfsCID
	if cid == "" {
		cid, _ = d.readSidecar(ref)
	}
	if cid == "" {
		return nil, err
	}
	cipher, err := d.ipfs.Cat(ctx, cid)
	if err != nil {
		return nil, ErrNotFound
	}
	return decrypt(password, cipher)
}

func (d *DualBackup) readCipher(ref string) ([]byte, error) {
	path := filepath.Join(d.disk.root, ref)
	return os.ReadFile(path)
}

type sidecar struct {
	IPFSCID string `json:"ipfsCid"`
}

func (d *DualBackup) writeSidecar(ref, cid string) error {
	path := filepath.Join(d.disk.root, ref) + ".ipfs.json"
	raw, _ := json.Marshal(sidecar{IPFSCID: cid})
	return os.WriteFile(path, raw, 0o600)
}

func (d *DualBackup) readSidecar(ref string) (string, error) {
	path := filepath.Join(d.disk.root, ref) + ".ipfs.json"
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	var s sidecar
	if err := json.Unmarshal(raw, &s); err != nil {
		return "", err
	}
	return s.IPFSCID, nil
}

// Disk exposes underlying disk store for tests.
func (d *DualBackup) Disk() *DiskBackup {
	return d.disk
}
