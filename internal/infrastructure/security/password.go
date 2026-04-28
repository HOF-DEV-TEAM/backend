package security

import (
	"crypto/md5" //nolint:gosec // MD5 is used only for legacy password migration, not for new passwords
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 12

// HashPassword returns a bcrypt hash of the plaintext password.
func HashPassword(plaintext string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plaintext), bcryptCost)
	if err != nil {
		return "", fmt.Errorf("hashing password: %w", err)
	}
	return string(b), nil
}

// CheckPasswordBcrypt reports whether the plaintext matches a bcrypt hash.
func CheckPasswordBcrypt(hashed, plaintext string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plaintext))
}

// MD5Hash returns the hexadecimal MD5 digest of the input.
// Used only for backwards-compatible login of legacy accounts.
func MD5Hash(plaintext string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(plaintext))) //nolint:gosec // G401: MD5 is intentional here for legacy password comparison only
}
