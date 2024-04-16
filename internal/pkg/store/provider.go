package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/Masterminds/squirrel"
	"github.com/ougirez/diplom/internal/domain"
	"github.com/ougirez/diplom/internal/domain/dto"
	"strings"
)

type ProviderStore interface {
	Insert(ctx context.Context, regionItem *dto.ProviderDto) (*domain.Provider, error)
}

var (
	regionsColumns         = "id, region_name, district_name, created_at, updated_at"
	providersColumns       = "id, region_id, provider_name, created_at, updated_at"
	groupedCategoryColumns = "id, region_id, name, created_at, updated_at"
	categoryColumns        = "id, grouped_category_id, name, unit, year_data, created_at, updated_at"
)

func (s *store) Insert(ctx context.Context, providerDto *dto.ProviderDto) (*domain.Provider, error) {
	region, err := s.GetRegionByName(ctx, providerDto.RegionName)
	if err != nil {
		return nil, fmt.Errorf("GetRegionByName: %w", err)
	}

	item, err := s.insertProvider(ctx, region.ID, providerDto)
	if err != nil {
		return nil, fmt.Errorf("insertProvider: %w", err)
	}

	for groupedCategoryName, grCategoryDto := range providerDto.GroupedCategories {
		groupedCategory, err := s.insertGroupedCategory(ctx, item.ID, groupedCategoryName)
		if err != nil {
			return nil, fmt.Errorf("insertGroupedCategory, region_name-%s, grouped_category_name-%s: %w", providerDto.RegionName, groupedCategoryName, err)
		}

		err = s.insertCategories(ctx, groupedCategory.ID, grCategoryDto.Categories)
		if err != nil {
			return nil, fmt.Errorf("insertCategories, region_name-%s, grouped_category_name-%s: %w", providerDto.RegionName, groupedCategoryName, err)
		}
	}

	return item, nil
}

func (s *store) insertProvider(ctx context.Context, regionID int64, regionItem *dto.ProviderDto) (*domain.Provider, error) {
	query := builder().Insert(tableRegions).
		Columns("id", "region_id", "provider_name").
		Values(regionItem.ProviderID, regionID, regionItem.ProviderName).
		Suffix(`on conflict (id) do nothing`)

	_, err := s.pool.Execx(ctx, query)
	if err != nil {
		return nil, err
	}

	selectQuery := builder().Select(strings.Split(providersColumns, ", ")...).
		From(tableRegions).
		Where(squirrel.Eq{"region_name": regionItem.RegionName})

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

	selectQuery := builder().Select(strings.Split(groupedCategoryColumns, ", ")...).
		From(tableGroupedCategories).
		Where(squirrel.And{
			squirrel.Eq{"provider_id": providerID},
			squirrel.Eq{"name": groupedCategoryName},
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

		query = query.Values(groupedCategoryID, categoryName, yearDataJSON)
	}

	query = query.Suffix(`
on conflict (group_id, name) 
do update
set year_data = excluded.year_data`)

	if _, err := s.pool.Execx(ctx, query); err != nil {
		return err
	}

	return nil
}
