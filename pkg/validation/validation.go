package validation

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"
)

// ValidateURL - проверяет валидность ссылки
func ValidateURL(key, expire, signature string) (bool, error) {

	if expiresInt, _ := strconv.Atoi(expire); time.Now().Unix() > int64(expiresInt) {
		return false, errors.New("Invalid expires")
	}
	if verifySignature(key, expire, signature) {
		return true, nil
	} else {
		return false, errors.New("Invalid signature")
	}
}

// verifySignature - проверяет подпись ссылки
func verifySignature(key, expires, receiveSignature string) bool {
	data := fmt.Sprintf("%s:%s", key, expires)
	secretKey := "adsfdsaffwe23r32ffew" // os.Getenv("HMAC_SECRET_KEY")

	h := hmac.New(sha256.New, []byte(secretKey))
	h.Write([]byte(data))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return hmac.Equal([]byte(expectedSignature), []byte(receiveSignature))
}
