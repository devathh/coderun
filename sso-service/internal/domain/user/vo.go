package user

import (
	"fmt"
	"net/mail"
	"strings"

	customerrors "github.com/devathh/coderun/sso-service/pkg/errors"
	"golang.org/x/crypto/bcrypt"
)

type Email string

func NewEmail(email string) (Email, error) {
	if _, err := mail.ParseAddress(email); err != nil {
		return Email(""), customerrors.ErrInvalidEmail
	}

	return Email(email), nil
}

type Password string

func NewPassword(password string) (Password, error) {
	password = strings.TrimSpace(password)
	if len([]rune(password)) < 8 {
		return Password(""), customerrors.ErrInvalidPassword
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return Password(""), fmt.Errorf("failed to generate hash from password: %w", err)
	}

	return Password(string(hash)), nil
}

func (p Password) Check(password string) bool {
	return bcrypt.CompareHashAndPassword([]byte(p), []byte(password)) == nil
}
