package store

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/ougirez/diplom/internal/domain"
	"strings"
)

func (s *store) GetRegionByName(ctx context.Context, regionName string) (*domain.Region, error) {
	query := builder().Select(strings.Split(regionsColumns, ", ")...).
		From(tableRegions).
		Where(squirrel.Eq{"region_name": regionName})

	var selected domain.Region
	err := s.pool.Selectx(ctx, &selected, query)
	if err != nil {
		return nil, err
	}

	return &selected, nil
}
