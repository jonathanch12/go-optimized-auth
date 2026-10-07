package store

import (
	"go-auth-api/internal/auth"
	"go-auth-api/internal/model"
)

type UserStore struct {
	byUsername map[string]model.User
}

func NewUserStore() *UserStore {
	hash, _ := auth.HashPassword("test123#")
	seed := model.User{ID: 1, Name: "Test", Username: "test1", Password: hash}
	return &UserStore{
		byUsername: map[string]model.User{seed.Username: seed},
	}
}

func (s *UserStore) FindByUsername(username string) (model.User, bool) {
	u, ok := s.byUsername[username]
	return u, ok
}

func (s *UserStore) FindByID(id int) (model.User, bool) {
	for _, u := range s.byUsername {
		if u.ID == id {
			return u, true
		}
	}
	return model.User{}, false
}
