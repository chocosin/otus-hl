package storage

import (
	"github.com/chocosin/otus-hl/social/model"
	uuid "github.com/satori/go.uuid"
)

type Storage interface {
	LastUsernames() ([]string, error)
	InsertUser(user *model.User) error
	FindUserByUsername(username string) (*model.User, error)

	InsertToken(token uuid.UUID, userId uuid.UUID) error
	DeleteToken(id uuid.UUID) error
	GetUserByToken(token uuid.UUID) (*model.User, error)
}
