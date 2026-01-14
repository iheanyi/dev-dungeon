package web

import (
	"crypto/sha256"
	"encoding/base64"

	"golang.org/x/crypto/ssh"
)

// Supported SSH key types (all standard secure algorithms).
var supportedKeyTypes = map[string]bool{
	ssh.KeyAlgoED25519:    true, // Ed25519 - recommended
	ssh.KeyAlgoECDSA256:   true, // ECDSA P-256
	ssh.KeyAlgoECDSA384:   true, // ECDSA P-384
	ssh.KeyAlgoECDSA521:   true, // ECDSA P-521
	ssh.KeyAlgoRSA:        true, // RSA
	ssh.KeyAlgoSKED25519:  true, // Security Key Ed25519
	ssh.KeyAlgoSKECDSA256: true, // Security Key ECDSA
}

// parseAuthorizedKey parses an SSH authorized_keys format public key.
func parseAuthorizedKey(in []byte) (ssh.PublicKey, string, []string, []byte, error) {
	return ssh.ParseAuthorizedKey(in)
}

// fingerprintSHA256 returns the SHA256 fingerprint of an SSH public key.
func fingerprintSHA256(key ssh.PublicKey) string {
	hash := sha256.Sum256(key.Marshal())
	return "SHA256:" + base64.RawStdEncoding.EncodeToString(hash[:])
}

// isKeyTypeSupported checks if the key type is one of the supported algorithms.
func isKeyTypeSupported(key ssh.PublicKey) bool {
	return supportedKeyTypes[key.Type()]
}
