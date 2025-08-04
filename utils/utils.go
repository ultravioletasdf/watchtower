package utils

import (
	"crypto/rand"
	"encoding/hex"
	"math/big"
	"net/mail"
	"regexp"
	"strings"
)

const VERIFY_CODE_MAX int64 = 999_999

var usernameRegex = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9._-]*$`)

func IsEmailValid(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// Returns whethere the username is suitable (alphanumeric + -._)
//
// Does not do a length check
func IsUsernameValid(username string) bool {
	return usernameRegex.MatchString(username)
}

// Returns whether n is between min and max (inclusive)
func Between(n, min, max int) bool {
	return n >= min && n <= max
}

// Gets the field name from a UNIQUE constraint failed error
func ErrUniqueConstraintFieldName(err error) string {
	parts := strings.SplitN(err.Error(), ":", 2)
	if len(parts) < 2 {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func GenerateString(length int) (string, error) {
	b := make([]byte, length/2)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
func GenerateVerifyCode() (int32, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(VERIFY_CODE_MAX))
	return int32(n.Int64()), err
}
