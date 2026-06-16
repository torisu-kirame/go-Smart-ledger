package userstore

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/navnitms/go-identicon/pkg/identicon"
)

const identiconPNGSize = 256

var (
	identiconOnce sync.Once
	identiconGen  *identicon.Identicon
	identiconErr  error
)

func identiconGenerator() (*identicon.Identicon, error) {
	identiconOnce.Do(func() {
		identiconGen, identiconErr = identicon.New(identicon.WithSize(identiconPNGSize))
	})
	return identiconGen, identiconErr
}

// AvatarSeedFromUserID hashes the user ID (SHA-256 hex) for deterministic identicon input.
func AvatarSeedFromUserID(userID string) string {
	sum := sha256.Sum256([]byte(userID))
	return hex.EncodeToString(sum[:])
}

// EnsureDefaultAvatar writes {userID}.png identicon when no avatar file exists yet.
func EnsureDefaultAvatar(dir, userID string) error {
	if userID == "" {
		return fmt.Errorf("empty user id")
	}
	if _, _, err := ResolveAvatarFile(dir, userID); err == nil {
		return nil
	}
	gen, err := identiconGenerator()
	if err != nil {
		return err
	}
	seed := AvatarSeedFromUserID(userID)
	img, err := gen.Generate(seed)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	path := filepath.Join(dir, userID+".png")
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	if err := gen.SavePNG(img, f); err != nil {
		_ = os.Remove(path)
		return err
	}
	return nil
}
