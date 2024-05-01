package store

import (
	"context"
	"github.com/ougirez/diplom/internal/domain"
	"github.com/ougirez/diplom/internal/domain/dto"
	"github.com/ougirez/diplom/internal/pkg/store/xpgx"
)

type Pool = xpgx.Pool

type Store interface {
	Insert(ctx context.Context, regionItem *dto.ProviderDto) (*domain.Provider, error)
	ListRegions(ctx context.Context) ([]*domain.Region, error)
	ListCategoriesByRegionID(ctx context.Context, opts ListCategoriesByRegionIDOpts) ([]*domain.ExtendedCategory, error)
	GetCategoryDataByRegions(ctx context.Context, opts GetCategoryDataByRegionsOpts) ([]*domain.RegionCategoryData, error)
}

type store struct {
	pool Pool
}

func NewStore(pool Pool) Store {
	return &store{pool}
}
