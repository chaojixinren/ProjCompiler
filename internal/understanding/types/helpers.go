package types

import (
	"crypto/sha1"
	"encoding/hex"
	"path/filepath"
	"strings"
)

const SchemaVersion = "v0.1"

func NewID(parts ...string) string {
	sum := sha1.New()
	for _, part := range parts {
		if _, err := sum.Write([]byte(part)); err != nil {
			continue
		}
		_, _ = sum.Write([]byte{0})
	}
	return hex.EncodeToString(sum.Sum(nil))[:12]
}

func NormalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	return filepath.ToSlash(filepath.Clean(path))
}

func ConfidenceBandFor(score float64) ConfidenceBand {
	switch {
	case score >= 0.9:
		return ConfidenceBandHigh
	case score >= 0.6:
		return ConfidenceBandMedium
	default:
		return ConfidenceBandLow
	}
}
