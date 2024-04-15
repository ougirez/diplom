package controller

import (
	"github.com/ougirez/diplom/internal/pkg/store"
	regionService "github.com/ougirez/diplom/internal/service/region"
)

type Controller struct {
	store   store.Store
	service *regionService.Service
}

func NewController(store store.Store, service *regionService.Service) *Controller {
	return &Controller{store: store, service: service}
}
