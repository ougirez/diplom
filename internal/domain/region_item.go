package domain

import "time"

type Year = int
type YearData = map[Year]float64

type Region struct {
	ID           int64     `db:"id"`
	RegionName   string    `db:"region_name"`
	DistrictName string    `db:"district_name"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type Provider struct {
	ID           int64     `db:"id"`
	RegionID     int64     `db:"region_id"`
	ProviderName string    `db:"name"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type GroupedCategory struct {
	ID        int64     `db:"id"`
	RegionID  int64     `db:"provider_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Category struct {
	ID        int64     `db:"id"`
	GroupID   int64     `db:"group_id"`
	Name      string    `db:"name"`
	Unit      string    `db:"unit"`
	YearData  YearData  `db:"year_data"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type ExtendedCategory struct {
	Category
	GroupName    string `db:"group_name"`
	ProviderName string `db:"provider_name"`
}

type RegionCategoryData struct {
	Unit                string   `db:"unit"`
	RegionName          string   `db:"region_name"`
	GroupedCategoryName string   `db:"grouped_category_name"`
	CategoryName        string   `db:"category_name"`
	YearData            YearData `db:"year_data"`
}
