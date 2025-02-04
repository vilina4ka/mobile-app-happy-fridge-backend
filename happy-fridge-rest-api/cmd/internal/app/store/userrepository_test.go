package store_test

import (
	"github.com/stretchr/testify/assert"
	"happy-fridge-rest-api/cmd/internal/app/model"
	"happy-fridge-rest-api/cmd/internal/app/store"
	"testing"
)

func TestUserRepository_Create(t *testing.T) {
	s, teardowm := store.TestStore(t, databaseURL)
	defer teardowm("users")

	u, err := s.User().Create(model.TestUser(t))
	assert.NoError(t, err)
	assert.NotNil(t, u)
}

func TestUserRepository_FindByEmail(t *testing.T) {
	s, teardowm := store.TestStore(t, databaseURL)
	defer teardowm("users")

	email := "user@example.com"

	_, err := s.User().FindByEmail(email)
	assert.Error(t, err)

	u := model.TestUser(t)
	u.Email = email
	s.User().Create(u)
	u, err = s.User().FindByEmail(email)
	assert.NoError(t, err)
	assert.NotNil(t, u)
}

func TestUserRepository_FindByFridgeNumber(t *testing.T) {
	s, teardowm := store.TestStore(t, databaseURL)
	defer teardowm("users")

	fridgeNumber := "000000"
	_, err := s.User().FindByFridgeNumber(fridgeNumber)
	assert.Error(t, err)
}
