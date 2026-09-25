-- migrate:up
create table auth_state (
    id uuid primary key default gen_random_uuid(),
    state text not null,
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

-- migrate:down
drop table auth_state;
