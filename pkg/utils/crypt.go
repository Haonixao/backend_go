package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"

	"backend_go/pkg/format_errors"

	"golang.org/x/crypto/bcrypt"
)

func Encrypt(text string, key string) (string, error) {
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", format_errors.Wrap(err, "ошибка создания cipher")
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", format_errors.Wrap(err, "ошибка создания GCM")
	}
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", format_errors.Wrap(err, "ошибка генерации nonce")
	}
	ciphertext := aesGCM.Seal(nonce, nonce, []byte(text), nil)
	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func Decrypt(encryptedText string, key string) (string, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", format_errors.Wrap(err, "ошибка декодирования ciphertext")
	}
	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", format_errors.Wrap(err, "ошибка создания cipher")
	}
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", format_errors.Wrap(err, "ошибка создания GCM")
	}
	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", format_errors.Wrap(errors.New("ciphertext слишком короткий"), "")
	}
	nonce := ciphertext[:nonceSize]
	ciphertext = ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", format_errors.Wrap(err, "ошибка decrypt")
	}
	return string(plaintext), nil
}

func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", format_errors.Wrap(err, "ошибка хэширования пароля "+password)
	}
	return string(hashedBytes), nil
}

func CheckPasswordHash(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		return format_errors.Wrap(err, "ошибка проверки пароля "+password+"и хэша "+hash)
	}
	return nil
}
