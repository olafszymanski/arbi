package utils

import (
	"crypto/hmac"
	"crypto/sha256"
	"fmt"
)

func Signature(secret, params string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(params))
	return fmt.Sprintf("%x", mac.Sum(nil))
}
