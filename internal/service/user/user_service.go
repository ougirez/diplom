package user

import (
	"github.com/ougirez/diplom/internal/pkg/store"
)

type Service struct {
	store *store.Store
}

func NewUserService(store *store.Store) *Service {
	return &Service{store}
}
