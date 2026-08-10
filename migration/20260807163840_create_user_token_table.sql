-- migrate:up
create table user_token (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null,
    access_token text not null,
    refresh_token text not null,
    expired_at timestamptz not null,
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

-- migrate:down
drop table user_token;
