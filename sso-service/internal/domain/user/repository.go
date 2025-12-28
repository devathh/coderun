package user

import (
	"context"
)

type MongoRepository interface {
	Save(context.Context, *User) (*User, error)
	Update(context.Context, *User) error // With id
	GetByID(context.Context, string) (*User, error)
	GetByEmail(context.Context, string) (*User, error)
}
