-- migrate:up
drop table query_tool;

-- migrate:down
create table query_tool (
    id uuid primary key default gen_random_uuid(),
    user_id uuid not null references user_information(id),
    columns jsonb not null,
    param_name text not null,
    param_type text not null,
    param_description text not null,
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);
