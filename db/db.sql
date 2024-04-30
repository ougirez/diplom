CREATE UNLOGGED TABLE IF NOT EXISTS regions
(
    id            BIGSERIAL PRIMARY KEY,
    region_name   text UNIQUE not null,
    district_name text        not null,
    square_km     bigint      not null,
    created_at    timestamptz not null default now(),
    updated_at    timestamptz not null default now(),

    UNIQUE (region_name)
);

CREATE UNLOGGED TABLE IF NOT EXISTS providers
(
    id         bigint PRIMARY KEY,
    region_id  bigint,
    name       text UNIQUE,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),

    UNIQUE (name),
    foreign key (region_id) references regions (id)
);

CREATE UNLOGGED TABLE IF NOT EXISTS grouped_categories
(
    id          BIGSERIAL PRIMARY KEY,
    provider_id bigint      NOT NULL,
    name        text        NOT NULL,
    created_at  timestamptz not null default now(),
    updated_at  timestamptz not null default now(),

    UNIQUE (provider_id, name),
    foreign key (provider_id) references providers (id)
);

CREATE UNLOGGED TABLE IF NOT EXISTS categories
(
    id         BIGSERIAL PRIMARY KEY,
    group_id   bigint      NOT NULL,
    name       text        NOT NULL,
    unit       text        NOT NULL,
    year_data  jsonb       NOT NULL,
    created_at timestamptz not null default now(),
    updated_at timestamptz not null default now(),

    UNIQUE (group_id, name),
    foreign key (group_id) references grouped_categories (id)
);