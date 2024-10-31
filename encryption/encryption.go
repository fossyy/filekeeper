package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/fossyy/filekeeper/types"
	"github.com/fossyy/filekeeper/utils"
)

type aesEncryption struct{}

func NewAesEncryption() types.Encryption {
	return &aesEncryption{}
}

func (e *aesEncryption) Encrypt(enc string) (string, error) {
	secret := utils.Getenv("ENCRYPTION_SECRET")

	aesCipher, err := aes.NewCipher([]byte(secret))

	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	_, err = rand.Read(nonce)
	if err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(enc), nil)

	ciphertextBase64 := base64.StdEncoding.EncodeToString(ciphertext)

	return ciphertextBase64, nil
}

func (e *aesEncryption) Decrypt(dec string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(dec)
	if err != nil {
		return "", err
	}

	secret := utils.Getenv("ENCRYPTION_SECRET")
	if len(secret) != 16 && len(secret) != 24 && len(secret) != 32 {
		return "", fmt.Errorf("invalid SECRET length: must be 16, 24, or 32 bytes")
	}

	aesCipher, err := aes.NewCipher([]byte(secret))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(aesCipher)
	if err != nil {
		return "", err
	}

	nonceSize := gcm.NonceSize()
	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("message authentication failed: %v", err)
	}

	return string(plaintext), nil
}
