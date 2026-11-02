package faker

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
)

// DefaultStagingBcryptHash is the standard bcrypt hash for password "password123" with cost 10.
const DefaultStagingBcryptHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOhiMeqB7v1G8eM6Yt1PkW86i7D4rTfe6"

// HashConstant returns the specified constant string or the default staging bcrypt hash.
func HashConstant(customValue string) string {
	if customValue != "" {
		return customValue
	}
	return DefaultStagingBcryptHash
}

// HashSHA256 returns a deterministic salted SHA256 hex string.
func HashSHA256(salt, rawValue string) string {
	h := sha256.New()
	h.Write([]byte(salt))
	h.Write([]byte{0})
	h.Write([]byte(rawValue))
	return hex.EncodeToString(h.Sum(nil))
}

// HashMD5 returns a deterministic salted MD5 hex string.
func HashMD5(salt, rawValue string) string {
	h := md5.New()
	h.Write([]byte(salt))
	h.Write([]byte{0})
	h.Write([]byte(rawValue))
	return hex.EncodeToString(h.Sum(nil))
}
