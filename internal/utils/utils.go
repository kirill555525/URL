package utils

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"github.com/kirill555525/URL/cmd/config"
)

const shortURLLength = 6 // 8 символов в base64

func GenerateShortID() (string, error) {
	b := make([]byte, shortURLLength)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	res := base64.URLEncoding.EncodeToString(b)

	return res, nil

}

func CreateShortURL(shortID string) string {
	cfg := config.GetConfig()
	return fmt.Sprintf("%s/%s", cfg.BaseURL, shortID)
}
