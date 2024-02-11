CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

create table webhooks (
    id uuid not null primary key default uuid_generate_v4(),
    target text not null,
    payload bytea not null,
    created_at timestamptz not null default now(),
    published_at timestamptz,
    deleted_at timestamptz
);

create index if not exists created_published on webhooks(published_at nulls first, created_at asc);