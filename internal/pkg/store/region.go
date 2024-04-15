package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ougirez/diplom/internal/domain"
)

type RegionItemStore interface {
	Insert(ctx context.Context, regionItem *domain.RegionItem) error
}

var regionItemColumns = []string{"id", "region_name", "district_name", "group_categories"}

func (s *store) Insert(ctx context.Context, regionItem *domain.RegionItem) error {
	groupCategoriesJSON, err := json.Marshal(regionItem.GroupedCategories)
	if err != nil {
		return fmt.Errorf("failed to marshal group categories: %w", err)
	}

	query := builder().Insert(tableRegionItems).
		Columns(regionItemColumns...).
		Values(regionItem.ID, regionItem.RegionName, regionItem.DistrictName, groupCategoriesJSON).
		Suffix(`
on conflict (id)
do update 
set region_name = excluded.region_name,
	district_name = excluded.district_name,
	group_categories = excluded.group_categories
`)

	sql, args, err := query.ToSql()
	if err != nil {
		return err
	}

	_, err = s.pool.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}
	return nil
}
