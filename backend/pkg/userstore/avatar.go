package userstore

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

const maxAvatarBytes = 2 << 20 // 2MB

var allowedAvatarTypes = map[string]string{
	"image/jpeg": ".jpg",
	"image/png":  ".png",
	"image/webp": ".webp",
	"image/gif":  ".gif",
}

// SaveAvatar writes avatar file to dir as {userID}{ext}.
func SaveAvatar(dir, userID string, r io.Reader, contentType string) (string, error) {
	ext, ok := allowedAvatarTypes[contentType]
	if !ok {
		return "", fmt.Errorf("unsupported image type: %s", contentType)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	// remove old avatars for this user
	prefix := filepath.Join(dir, userID)
	matches, _ := filepath.Glob(prefix + ".*")
	for _, old := range matches {
		_ = os.Remove(old)
	}
	path := prefix + ext
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	limited := io.LimitReader(r, maxAvatarBytes+1)
	n, err := io.Copy(f, limited)
	if err != nil {
		_ = os.Remove(path)
		return "", err
	}
	if n > maxAvatarBytes {
		_ = os.Remove(path)
		return "", fmt.Errorf("avatar too large (max 2MB)")
	}
	return path, nil
}

// RemoveAvatarFiles deletes all avatar files for a user.
func RemoveAvatarFiles(dir, userID string) {
	matches, _ := filepath.Glob(filepath.Join(dir, userID) + ".*")
	for _, p := range matches {
		_ = os.Remove(p)
	}
}

// ResolveAvatarFile finds avatar file on disk for user.
func ResolveAvatarFile(dir, userID string) (path string, contentType string, err error) {
	for ct, ext := range allowedAvatarTypes {
		p := filepath.Join(dir, userID+ext)
		if _, err := os.Stat(p); err == nil {
			return p, ct, nil
		}
	}
	return "", "", os.ErrNotExist
}

// DetectImageType reads header bytes for content type.
func DetectImageType(r io.Reader) (string, io.Reader, error) {
	buf := make([]byte, 512)
	n, err := r.Read(buf)
	if err != nil && err != io.EOF {
		return "", nil, err
	}
	ct := http.DetectContentType(buf[:n])
	return ct, io.MultiReader(bytes.NewReader(buf[:n]), r), nil
}
