CREATE UNLOGGED TABLE IF NOT EXISTS region_items
(
    id            bigint PRIMARY KEY,
    region_name   text UNIQUE,
    district_name text,
    created_at    timestamptz not null default now(),
    updated_at    timestamptz not null default now()
);

CREATE UNLOGGED TABLE IF NOT EXISTS grouped_categories
(
    id         BIGSERIAL PRIMARY KEY,
    region_id  bigint      NOT NULL,
    name       text        NOT NULL,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),

    UNIQUE (region_id, name)
);

CREATE UNLOGGED TABLE IF NOT EXISTS categories
(
    id         BIGSERIAL PRIMARY KEY,
    grouped_id bigint      NOT NULL,
    name       text        NOT NULL,
    unit       text        NOT NULL,
    year_data  jsonb       NOT NULL,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),

    UNIQUE (grouped_id, name)
);