package dto

import (
	"fmt"
	"github.com/ougirez/diplom/internal/domain"
	"sync"
)

type Category struct {
	Unit       string
	YearData   map[domain.Year]float64
	yearDataMx sync.Mutex
}

func (c *Category) PutData(year domain.Year, data float64, unit string) error {
	c.yearDataMx.Lock()
	defer c.yearDataMx.Unlock()

	if c.Unit == "" {
		c.Unit = unit
	} else {
		if c.Unit != unit {
			return fmt.Errorf("different units for one cateogory: %s and %s", c.Unit, unit)
		}
	}

	c.YearData[year] = data
	return nil
}

type GroupedCategory struct {
	Categories   map[string]*Category `bson:"categories"`
	categoriesMx sync.Mutex
}

func (gc *GroupedCategory) PutCategory(categoryName string, category *Category) {
	gc.categoriesMx.Lock()
	defer gc.categoriesMx.Unlock()

	gc.Categories[categoryName] = category
}

func (gc *GroupedCategory) GetCategory(categoryName string) *Category {
	gc.categoriesMx.Lock()
	defer gc.categoriesMx.Unlock()

	category, ok := gc.Categories[categoryName]
	if !ok {
		category = &Category{
			YearData: make(map[domain.Year]float64),
		}
		gc.Categories[categoryName] = category
	}

	return category
}

type ProviderDto struct {
	ProviderID        int64
	DistrictName      string
	RegionName        string
	ProviderName      string
	GroupedCategories map[string]*GroupedCategory
	groupCategoriesMx sync.Mutex
}

func (r *ProviderDto) PutGroupCategory(groupedCategoryName string, groupCategory *GroupedCategory) {
	r.groupCategoriesMx.Lock()
	defer r.groupCategoriesMx.Unlock()

	r.GroupedCategories[groupedCategoryName] = groupCategory
}

func (r *ProviderDto) GetGroupCategory(groupedCategoryName string) *GroupedCategory {
	r.groupCategoriesMx.Lock()
	defer r.groupCategoriesMx.Unlock()

	groupedCategory, ok := r.GroupedCategories[groupedCategoryName]
	if !ok {
		groupedCategory = &GroupedCategory{
			Categories: make(map[string]*Category),
		}
		r.GroupedCategories[groupedCategoryName] = groupedCategory
	}

	return groupedCategory
}
