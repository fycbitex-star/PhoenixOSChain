package crypto

import (
	"crypto/ed25519"
	"encoding/hex"
)

func Sign(priv ed25519.PrivateKey, message []byte) string {
	return hex.EncodeToString(ed25519.Sign(priv, message))
}

func Verify(pub ed25519.PublicKey, message []byte, signatureHex string) bool {
	signature, err := hex.DecodeString(signatureHex)
	if err != nil {
		return false
	}
	return ed25519.Verify(pub, message, signature)
}
