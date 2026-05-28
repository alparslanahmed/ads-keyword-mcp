package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

// unitSeparator joins parts so distinct field boundaries cannot collide.
const unitSeparator = "\x1f"

// SHA256Hex returns the hex-encoded SHA-256 of the parts joined by a unit
// separator. Used for deterministic cache keys and request fingerprints.
func SHA256Hex(parts ...string) string {
	h := sha256.New()
	h.Write([]byte(strings.Join(parts, unitSeparator)))
	return hex.EncodeToString(h.Sum(nil))
}
