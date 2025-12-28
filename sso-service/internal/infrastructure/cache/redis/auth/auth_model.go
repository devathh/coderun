package authredis

import "github.com/google/uuid"

type SessionModel struct {
	UserID uuid.UUID `json:"user_id"`
	Email  string    `json:"email"`
}
