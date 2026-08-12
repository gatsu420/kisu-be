-- migrate:up
create table tool (
    id uuid primary key default gen_random_uuid(),
    tool_description text not null,
    table_name text not null,
    columns jsonb not null,
    query_examples jsonb not null,
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

-- migrate:down
drop table tool;
