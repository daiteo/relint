package userstore

import (
	"database/sql"
	"errors"
	gorm "github.com/daiteo/relint/example/src/lint021/gormstub"
)

type UserStore struct{}

func (s *UserStore) GetUser() error {
	return sql.ErrNoRows // want `LINT-021: store function must not return sql\.ErrNoRows directly, wrap it in a domain error`
}

func (s *UserStore) FindUser() error {
	err := sql.ErrNoRows
	return errors.New("wrapped: " + err.Error()) // ok - not returning ErrNoRows directly
}

func (s *UserStore) FindUserWithGorm() error {
	return gorm.ErrRecordNotFound // want `LINT-021: store function must not return gorm\.ErrRecordNotFound directly, wrap it in a domain error`
}
