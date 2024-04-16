package domain

import "time"

type Year = int

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
	ProviderName string    `db:"provider_name"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}

type GroupedCategory struct {
	ID        int64     `db:"id"`
	RegionID  int64     `db:"region_id"`
	Name      string    `db:"name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type Category struct {
	ID        int64            `db:"id"`
	GroupID   int64            `db:"group_id"`
	Name      string           `db:"name"`
	Unit      string           `db:"unit"`
	YearData  map[Year]float64 `db:"year_data"`
	CreatedAt time.Time        `db:"created_at"`
	UpdatedAt time.Time        `db:"updated_at"`
}
