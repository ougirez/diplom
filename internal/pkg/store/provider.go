package store

import (
	"context"
	"encoding/json"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"github.com/ougirez/diplom/internal/domain"
	"github.com/ougirez/diplom/internal/domain/dto"
	"github.com/ougirez/diplom/internal/pkg/logger"
)

type ListCategoriesByRegionIDOpts struct {
	RegionID          int
	CategoryName      *string
	GroupCategoryName *string
}

type ProviderStore interface {
	Insert(ctx context.Context, regionItem *dto.ProviderDto) (*domain.Provider, error)
	ListRegions(ctx context.Context) ([]*domain.Region, error)
	ListCategoriesByRegionID(ctx context.Context, opts ListCategoriesByRegionIDOpts) ([]*domain.ExtendedCategory, error)
}

var (
	regionsColumns         = []string{"id", "region_name", "district_name", "created_at", "updated_at"}
	providersColumns       = []string{"id", "region_id", "name", "created_at", "updated_at"}
	groupedCategoryColumns = []string{"id", "provider_id", "name", "created_at", "updated_at"}
	categoryColumns        = []string{"id", "grouped_category_id", "name", "unit", "year_data", "created_at", "updated_at"}
)

func (s *store) Insert(ctx context.Context, providerDto *dto.ProviderDto) (*domain.Provider, error) {
	region, err := s.insertRegion(ctx, providerDto.RegionName, providerDto.DistrictName)
	if err != nil {
		logger.Errorf(ctx, "insertRegion: %s", err.Error())
		return nil, fmt.Errorf("insertRegion: %w", err)
	}

	item, err := s.insertProvider(ctx, region.ID, providerDto)
	if err != nil {
		logger.Errorf(ctx, "insertProvider: %s", err.Error())
		return nil, fmt.Errorf("insertProvider: %w", err)
	}

	for groupedCategoryName, grCategoryDto := range providerDto.GroupedCategories {
		groupedCategory, err := s.insertGroupedCategory(ctx, item.ID, groupedCategoryName)
		if err != nil {
			logger.Errorf(ctx, "insertGroupedCategory: %s", err.Error())
			return nil, fmt.Errorf("insertGroupedCategory, region_name-%s, grouped_category_name-%s: %w", providerDto.RegionName, groupedCategoryName, err)
		}

		err = s.insertCategories(ctx, groupedCategory.ID, grCategoryDto.Categories)
		if err != nil {
			logger.Errorf(ctx, "insertCategories: %s", err.Error())
			return nil, fmt.Errorf("insertCategories, region_name-%s, grouped_category_name-%s: %w", providerDto.RegionName, groupedCategoryName, err)
		}
	}

	return item, nil
}

func (s *store) insertRegion(ctx context.Context, regionName string, districtName string) (*domain.Region, error) {
	query := builder().Insert(tableRegions).
		Columns("region_name", "district_name").
		Values(regionName, districtName).
		Suffix(`on conflict (region_name) do update set district_name=excluded.district_name`)

	_, err := s.pool.Execx(ctx, query)
	if err != nil {
		return nil, err
	}

	selectQuery := builder().Select(regionsColumns...).
		From(tableRegions).
		Where(sq.Eq{"region_name": regionName})

	var selected domain.Region
	err = s.pool.Getx(ctx, &selected, selectQuery)
	if err != nil {
		return nil, err
	}

	return &selected, nil
}

func (s *store) insertProvider(ctx context.Context, regionID int64, regionItem *dto.ProviderDto) (*domain.Provider, error) {
	query := builder().Insert(tableProviders).
		Columns("id", "region_id", "name").
		Values(regionItem.ProviderID, regionID, regionItem.ProviderName).
		Suffix(`on conflict (id) do nothing`)

	_, err := s.pool.Execx(ctx, query)
	if err != nil {
		return nil, err
	}

	selectQuery := builder().Select(providersColumns...).
		From(tableProviders).
		Where(sq.Eq{"name": regionItem.ProviderName})

	var selected domain.Provider
	err = s.pool.Getx(ctx, &selected, selectQuery)
	if err != nil {
		return nil, err
	}

	return &selected, nil
}

func (s *store) insertGroupedCategory(
	ctx context.Context,
	providerID int64,
	groupedCategoryName string,
) (*domain.GroupedCategory, error) {
	query := builder().Insert(tableGroupedCategories).
		Columns("provider_id", "name").
		Values(providerID, groupedCategoryName).
		Suffix("on conflict do nothing")

	_, err := s.pool.Execx(ctx, query)
	if err != nil {
		return nil, err
	}

	selectQuery := builder().Select(groupedCategoryColumns...).
		From(tableGroupedCategories).
		Where(sq.And{
			sq.Eq{"provider_id": providerID},
			sq.Eq{"name": groupedCategoryName},
		})

	var selected domain.GroupedCategory
	err = s.pool.Getx(ctx, &selected, selectQuery)
	if err != nil {
		return nil, err
	}

	return &selected, nil
}

func (s *store) insertCategories(
	ctx context.Context,
	groupedCategoryID int64,
	categoryDtos map[string]*dto.Category,
) error {

	query := builder().Insert(tableCategories).
		Columns("group_id", "name", "unit", "year_data")

	for categoryName, categoryDto := range categoryDtos {
		yearDataJSON, err := json.Marshal(categoryDto.YearData)
		if err != nil {
			return fmt.Errorf("failed to marshal year data: %w", err)
		}

		query = query.Values(groupedCategoryID, categoryName, categoryDto.Unit, yearDataJSON)
	}

	query = query.Suffix(`
on conflict (group_id, name) 
do update
set 
	year_data = excluded.year_data,
	unit = excluded.unit`)

	if _, err := s.pool.Execx(ctx, query); err != nil {
		logger.Error(ctx, err.Error())
		return err
	}

	return nil
}

func (s *store) ListRegions(ctx context.Context) ([]*domain.Region, error) {
	query := builder().Select(regionsColumns...).
		From(tableRegions).
		OrderBy("district_name, region_name")

	var selected []*domain.Region

	err := s.pool.Selectx(ctx, &selected, query)
	if err != nil {
		return nil, err
	}

	return selected, nil
}

func (s *store) ListCategoriesByRegionID(
	ctx context.Context,
	opts ListCategoriesByRegionIDOpts,
) ([]*domain.ExtendedCategory, error) {
	query := builder().Select(
		`c.*, gc.name as group_name, p.name as provider_name`).
		From("grouped_categories gc").
		Join("categories c on gc.id=c.group_id").
		Join("providers p on p.id=gc.provider_id").
		Join("regions r on r.id=p.region_id").
		Where(sq.Eq{"r.id": opts.RegionID})

	if opts.CategoryName != nil {
		query = query.Where(sq.Eq{"c.name": *opts.CategoryName})
	}

	if opts.GroupCategoryName != nil {
		query = query.Where(sq.Eq{"gc.name": *opts.GroupCategoryName})
	}

	var selected []*domain.ExtendedCategory

	err := s.pool.Selectx(ctx, &selected, query)
	if err != nil {
		logger.Error(ctx, err.Error())
		return nil, err
	}

	return selected, nil
}
