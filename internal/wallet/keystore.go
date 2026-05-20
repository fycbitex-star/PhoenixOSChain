package wallet

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/phoenixchain/phoenixchain/internal/crypto"
)

const kdfRounds = 120000

type KeystoreFile struct {
	Version    int    `json:"version"`
	Address    string `json:"address"`
	PublicKey  string `json:"public_key"`
	KDF        string `json:"kdf"`
	KDFRounds  int    `json:"kdf_rounds"`
	Salt       string `json:"salt"`
	Nonce      string `json:"nonce"`
	CipherText string `json:"cipher_text"`
}

type Wallet struct {
	Address    string
	PublicKey  ed25519.PublicKey
	PrivateKey ed25519.PrivateKey
}

func New() (Wallet, error) {
	pub, priv, err := crypto.GenerateKeypair()
	if err != nil {
		return Wallet{}, err
	}
	return Wallet{Address: crypto.AddressFromPublicKey(pub), PublicKey: pub, PrivateKey: priv}, nil
}

func SaveEncrypted(path string, w Wallet, password string) error {
	if password == "" {
		return fmt.Errorf("password is required")
	}
	salt := randomBytes(16)
	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return err
	}
	nonce := randomBytes(gcm.NonceSize())
	cipherText := gcm.Seal(nil, nonce, w.PrivateKey, []byte(w.Address))
	file := KeystoreFile{
		Version:    1,
		Address:    w.Address,
		PublicKey:  crypto.EncodePublicKey(w.PublicKey),
		KDF:        "phoenix-pbkdf2-sha256",
		KDFRounds:  kdfRounds,
		Salt:       hex.EncodeToString(salt),
		Nonce:      hex.EncodeToString(nonce),
		CipherText: hex.EncodeToString(cipherText),
	}
	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

func LoadEncrypted(path string, password string) (Wallet, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Wallet{}, err
	}
	var file KeystoreFile
	if err := json.Unmarshal(data, &file); err != nil {
		return Wallet{}, err
	}
	salt, err := hex.DecodeString(file.Salt)
	if err != nil {
		return Wallet{}, err
	}
	nonce, err := hex.DecodeString(file.Nonce)
	if err != nil {
		return Wallet{}, err
	}
	cipherText, err := hex.DecodeString(file.CipherText)
	if err != nil {
		return Wallet{}, err
	}
	key := deriveKey(password, salt)
	block, err := aes.NewCipher(key)
	if err != nil {
		return Wallet{}, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return Wallet{}, err
	}
	plain, err := gcm.Open(nil, nonce, cipherText, []byte(file.Address))
	if err != nil {
		return Wallet{}, fmt.Errorf("decrypt keystore: %w", err)
	}
	priv, err := crypto.DecodePrivateKey(hex.EncodeToString(plain))
	if err != nil {
		return Wallet{}, err
	}
	pub, err := crypto.DecodePublicKey(file.PublicKey)
	if err != nil {
		return Wallet{}, err
	}
	if crypto.AddressFromPublicKey(pub) != file.Address {
		return Wallet{}, fmt.Errorf("address does not match public key")
	}
	return Wallet{Address: file.Address, PublicKey: pub, PrivateKey: priv}, nil
}

func randomBytes(n int) []byte {
	data := make([]byte, n)
	if _, err := rand.Read(data); err != nil {
		panic(err)
	}
	return data
}

func deriveKey(password string, salt []byte) []byte {
	block := append([]byte(password), salt...)
	sum := sha256.Sum256(block)
	key := sum[:]
	for i := 0; i < kdfRounds; i++ {
		round := sha256.Sum256(append(key, salt...))
		key = round[:]
	}
	out := make([]byte, 32)
	copy(out, key)
	return out
}
