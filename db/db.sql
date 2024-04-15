CREATE UNLOGGED TABLE IF NOT EXISTS region_items
(
    "id"               text PRIMARY KEY,
    "region_name"      text UNIQUE,
    "district_name"    text,
    "group_categories" jsonb,
    created_at         timestamptz not null default now(),
    updated_at         timestamptz not null default now()
);