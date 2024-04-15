package domain

import (
	"fmt"
	"sync"
)

type Year = int

type Category struct {
	Unit       string           `bson:"unit"`
	YearData   map[Year]float64 `bson:"year_data"`
	yearDataMx sync.Mutex
}

func (c *Category) PutData(year Year, data float64, unit string) error {
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
			YearData: make(map[Year]float64),
		}
		gc.Categories[categoryName] = category
	}

	return category
}

type RegionItem struct {
	ID                string                      `bson:"_id"`
	DistrictName      string                      `bson:"district_name"`
	RegionName        string                      `bson:"region_name"`
	GroupedCategories map[string]*GroupedCategory `bson:"grouped_categories"`
	groupCategoriesMx sync.Mutex
}

func (r *RegionItem) PutGroupCategory(groupedCategoryName string, groupCategory *GroupedCategory) {
	r.groupCategoriesMx.Lock()
	defer r.groupCategoriesMx.Unlock()

	r.GroupedCategories[groupedCategoryName] = groupCategory
}

func (r *RegionItem) GetGroupCategory(groupedCategoryName string) *GroupedCategory {
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
