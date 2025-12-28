package user

import (
	"errors"
	"testing"

	customerrors "github.com/devathh/coderun/sso-service/pkg/errors"
)

type input struct {
	username string
	email    string
	password string
}

func TestCreatingUser(t *testing.T) {
	testCases := []struct {
		Name    string
		Input   input
		WantErr error
	}{
		{Name: "base", Input: input{
			username: "example_username",
			email:    "mail@example.com",
			password: "very_secret_password",
		}, WantErr: nil},

		{Name: "invalid_username", Input: input{
			username: "",
			email:    "mail@example.com",
			password: "very_secret_password",
		}, WantErr: customerrors.ErrInvalidUsername},

		{Name: "invalid_password", Input: input{
			username: "example_username",
			email:    "mail@example.com",
			password: "",
		}, WantErr: customerrors.ErrInvalidPassword},

		{Name: "invalid_email", Input: input{
			username: "example_username",
			email:    "mail",
			password: "very_secret_password",
		}, WantErr: customerrors.ErrInvalidEmail},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			email, err := NewEmail(tc.Input.email)
			if err != nil && !errors.Is(tc.WantErr, err) {
				t.Errorf("got %v, want %v", err, tc.WantErr)
			}

			password, err := NewPassword(tc.Input.password)
			if err != nil && !errors.Is(tc.WantErr, err) {
				t.Errorf("got %v, want %v", err, tc.WantErr)
			}

			_, err = New(tc.Input.username, email, password)
			if err != nil && !errors.Is(tc.WantErr, err) {
				t.Errorf("got %v, want %v", err, tc.WantErr)
			}
		})
	}
}
