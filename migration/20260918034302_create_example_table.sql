-- migrate:up
create table example (
    id uuid primary key default gen_random_uuid(),
    tool_id uuid not null references tool(id),
    query text not null,
    description text not null,
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

-- migrate:down
drop table example;
