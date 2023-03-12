package user

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"time"
)

func EncodeString(s string) string {
	data := base64.StdEncoding.EncodeToString([]byte(s))
	return string(data)
}

func DecodeString(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

type OtpGenerator interface {
	Generate() string
}

type otpGenerator struct{}

func New() OtpGenerator {
	rand.Seed(time.Now().UnixNano())
	return &otpGenerator{}
}

func (*otpGenerator) Generate() string {
	return fmt.Sprintf("%06d", rand.Intn(999999))
}
