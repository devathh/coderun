package usermongo

import (
	"github.com/devathh/coderun/sso-service/internal/domain/user"
	"github.com/google/uuid"
)

func toModel(user *user.User) UserModel {
	return UserModel{
		ID:           user.ID().String(),
		Username:     user.Username(),
		Email:        string(user.Email()),
		PasswordHash: string(user.Password()),
	}
}

func toDomain(model *UserModel) (*user.User, error) {
	email, err := user.NewEmail(model.Email)
	if err != nil {
		return nil, err
	}

	passwordHash := user.Password(model.PasswordHash)

	id, err := uuid.Parse(model.ID)
	if err != nil {
		return nil, err
	}

	user := user.From(id, model.Username, email, passwordHash)

	return user, nil
}
