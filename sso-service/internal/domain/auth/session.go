package auth

import (
	"net/mail"
	"strings"

	customerrors "github.com/devathh/coderun/sso-service/pkg/errors"
	"github.com/google/uuid"
)

type Session struct {
	userID uuid.UUID
	email  string
}

func NewSession(userID uuid.UUID, email string) (*Session, error) {
	email = strings.TrimSpace(email)
	if _, err := mail.ParseAddress(email); err != nil {
		return nil, customerrors.ErrInvalidEmail
	}

	return &Session{
		userID: userID,
		email:  email,
	}, nil
}

func (s *Session) Email() string {
	return s.email
}

func (s *Session) UserID() uuid.UUID {
	return s.userID
}
