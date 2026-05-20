package wallet

import (
	"crypto/rand"
	"encoding/hex"
)

func GenerateMnemonic() (string, error) {
	data := make([]byte, 16)
	if _, err := rand.Read(data); err != nil {
		return "", err
	}
	return hex.EncodeToString(data), nil
}
