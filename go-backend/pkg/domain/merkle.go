package domain

import (
	"crypto/sha256"
	"encoding/hex"
)

func MerkleRoot(hashes []string) string {
	if len(hashes) == 0 {
		sum := sha256.Sum256(nil)
		return hex.EncodeToString(sum[:])
	}
	layer := make([][]byte, 0, len(hashes))
	for _, h := range hashes {
		b, err := hex.DecodeString(h)
		if err != nil || len(b) == 0 {
			sum := sha256.Sum256([]byte(h))
			b = sum[:]
		}
		layer = append(layer, b)
	}
	for len(layer) > 1 {
		next := make([][]byte, 0, (len(layer)+1)/2)
		for i := 0; i < len(layer); i += 2 {
			if i+1 < len(layer) {
				next = append(next, hashPair(layer[i], layer[i+1]))
			} else {
				next = append(next, hashPair(layer[i], layer[i]))
			}
		}
		layer = next
	}
	return hex.EncodeToString(layer[0])
}

func EventHash(ledgerID string, seq uint64, eventType string, payload []byte, prevHash string) string {
	h := sha256.New()
	h.Write([]byte(ledgerID))
	var buf [8]byte
	for i := 0; i < 8; i++ {
		buf[7-i] = byte(seq >> (8 * i))
	}
	h.Write(buf[:])
	h.Write([]byte(eventType))
	h.Write(payload)
	h.Write([]byte(prevHash))
	return hex.EncodeToString(h.Sum(nil))
}

func hashPair(a, b []byte) []byte {
	h := sha256.New()
	h.Write(a)
	h.Write(b)
	return h.Sum(nil)
}
