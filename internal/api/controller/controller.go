package controller

import (
	regionService "github.com/ougirez/diplom/internal/service/providers"
)

type Controller struct {
	service *regionService.Service
}

func NewController(service *regionService.Service) *Controller {
	return &Controller{service: service}
}
