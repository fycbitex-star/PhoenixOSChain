package crypto

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func GenerateKeypair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	return ed25519.GenerateKey(rand.Reader)
}

func AddressFromPublicKey(pub ed25519.PublicKey) string {
	hash := HashBytes(pub)
	return "phx" + hash[:40]
}

func EncodePrivateKey(priv ed25519.PrivateKey) string {
	return hex.EncodeToString(priv)
}

func DecodePrivateKey(value string) (ed25519.PrivateKey, error) {
	raw, err := hex.DecodeString(value)
	if err != nil {
		return nil, err
	}
	if len(raw) != ed25519.PrivateKeySize {
		return nil, fmt.Errorf("invalid private key length")
	}
	return ed25519.PrivateKey(raw), nil
}

func EncodePublicKey(pub ed25519.PublicKey) string {
	return hex.EncodeToString(pub)
}

func DecodePublicKey(value string) (ed25519.PublicKey, error) {
	raw, err := hex.DecodeString(value)
	if err != nil {
		return nil, err
	}
	if len(raw) != ed25519.PublicKeySize {
		return nil, fmt.Errorf("invalid public key length")
	}
	return ed25519.PublicKey(raw), nil
}
