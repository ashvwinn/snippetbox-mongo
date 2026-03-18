package mocks

import "github.com/ASH-WIN-10/snippetbox/internal/models"

var mockUserID = "65a1b2c3d4e5f67890abcdef"

type UserModel struct{}

func (m *UserModel) Insert(name, email, password string) error {
	switch email {
	case "dupe@example.com":
		return models.ErrDuplicateEmail
	default:
		return nil
	}
}

func (m *UserModel) Authenticate(email, password string) (string, error) {
	if email == "alice@example.com" && password == "pa$$word" {
		return mockUserID, nil
	}

	return "", models.ErrInvalidCredentials
}

func (m *UserModel) Exists(id string) (bool, error) {
	switch id {
	case mockUserID:
		return true, nil
	default:
		return false, nil
	}
}
