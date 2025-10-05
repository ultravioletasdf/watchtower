package utils

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"net/mail"
	"regexp"
	"strings"
	"time"
	common "videoapp/internal/errors"

	"github.com/jackc/pgx/v5/pgtype"
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

func PgTextFromPointer(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

type PresignedToken struct {
	Prefix   string
	ExpireAt time.Time
}

// Takes an upload id and generates a URL for accessing that video with a prefix, expiry and trustable signature
func GeneratePresignedUrl(privateKey *ecdsa.PrivateKey, uploadId int64, session string) (string, string, error) {
	token := PresignedToken{Prefix: fmt.Sprintf("/videos/%d", uploadId), ExpireAt: time.Now().Add(time.Hour * 6)}
	data, err := json.Marshal(token)
	if err != nil {
		return "", "", common.ErrInternal(err)
	}
	hash := sha256.Sum256([]byte(data))

	sigBytes, err := ecdsa.SignASN1(rand.Reader, privateKey, hash[:])
	if err != nil {
		return "", "", common.ErrInternal(err)
	}
	payload := base64.RawURLEncoding.EncodeToString(data)
	sig := base64.RawURLEncoding.EncodeToString(sigBytes)
	return payload, sig, nil
}
