package user

import (
	"strings"

	customerrors "github.com/devathh/coderun/sso-service/pkg/errors"
	"github.com/google/uuid"
)

type User struct {
	id       uuid.UUID
	email    Email
	username string
	password Password
}

func New(username string, email Email, password Password) (*User, error) {
	username = strings.TrimSpace(username)
	if username == "" {
		return nil, customerrors.ErrInvalidUsername
	}

	return &User{
		id:       uuid.New(),
		email:    email,
		username: username,
		password: password,
	}, nil
}

func From(id uuid.UUID, username string, email Email, password Password) *User {
	return &User{
		id:       id,
		username: username,
		email:    email,
		password: password,
	}
}

func (u *User) ID() uuid.UUID {
	return u.id
}

func (u *User) Email() Email {
	return u.email
}

func (u *User) Username() string {
	return u.username
}

func (u *User) Password() Password {
	return u.password
}
