-- migrate:up
create table tool_param (
    id uuid primary key default gen_random_uuid(),
    tool_id uuid not null references tool(id),
    name text not null,
    type text not null,
    description text not null,
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

-- migrate:down
drop table tool_param;
