package tx

import (
	"crypto/ed25519"

	phxcrypto "github.com/phoenixchain/phoenixchain/internal/crypto"
)

func AddressForPrivateKey(priv ed25519.PrivateKey) string {
	return phxcrypto.AddressFromPublicKey(priv.Public().(ed25519.PublicKey))
}
