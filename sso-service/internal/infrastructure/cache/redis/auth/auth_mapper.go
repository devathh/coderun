package authredis

import "github.com/devathh/coderun/sso-service/internal/domain/auth"

func toModel(session *auth.Session) *SessionModel {
	return &SessionModel{
		UserID: session.UserID(),
		Email:  session.Email(),
	}
}

func toDomain(model *SessionModel) (*auth.Session, error) {
	session, err := auth.NewSession(model.UserID, model.Email)
	if err != nil {
		return nil, err
	}

	return session, nil
}
