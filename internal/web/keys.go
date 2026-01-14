package web

import (
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/ssh"
)

// parseAuthorizedKey parses an SSH authorized_keys format public key.
func parseAuthorizedKey(in []byte) (ssh.PublicKey, string, []string, []byte, error) {
	return ssh.ParseAuthorizedKey(in)
}

// fingerprintSHA256 returns the SHA256 fingerprint of an SSH public key.
func fingerprintSHA256(key ssh.PublicKey) string {
	hash := sha256.Sum256(key.Marshal())
	return "SHA256:" + base64.RawStdEncoding.EncodeToString(hash[:])
}
