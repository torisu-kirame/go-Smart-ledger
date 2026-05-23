package userstore

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAvatarSeedFromUserID_deterministic(t *testing.T) {
	a := AvatarSeedFromUserID("1234567890123456789")
	b := AvatarSeedFromUserID("1234567890123456789")
	if a != b || len(a) != 64 {
		t.Fatalf("unexpected seed: %q len=%d", a, len(a))
	}
	if AvatarSeedFromUserID("other") == a {
		t.Fatal("different users should have different seeds")
	}
}

func TestEnsureDefaultAvatar_writesPNG(t *testing.T) {
	dir := t.TempDir()
	userID := "9876543210987654321"
	if err := EnsureDefaultAvatar(dir, userID); err != nil {
		t.Fatal(err)
	}
	path, ct, err := ResolveAvatarFile(dir, userID)
	if err != nil {
		t.Fatal(err)
	}
	if ct != "image/png" {
		t.Fatalf("content type = %q", ct)
	}
	info, err := os.Stat(path)
	if err != nil || info.Size() == 0 {
		t.Fatalf("stat: %v size=%d", err, info.Size())
	}
	// idempotent
	if err := EnsureDefaultAvatar(dir, userID); err != nil {
		t.Fatal(err)
	}
	if got, _, _ := ResolveAvatarFile(dir, userID); got != path {
		t.Fatalf("second call changed path: %s -> %s", path, got)
	}
}

func TestEnsureDefaultAvatar_respectsUploadedAvatar(t *testing.T) {
	dir := t.TempDir()
	userID := "111"
	custom := filepath.Join(dir, userID+".jpg")
	if err := os.WriteFile(custom, []byte{0xff, 0xd8, 0xff}, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureDefaultAvatar(dir, userID); err != nil {
		t.Fatal(err)
	}
	path, _, err := ResolveAvatarFile(dir, userID)
	if err != nil {
		t.Fatal(err)
	}
	if path != custom {
		t.Fatalf("expected custom avatar %s, got %s", custom, path)
	}
}
