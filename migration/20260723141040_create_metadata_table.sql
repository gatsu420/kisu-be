-- migrate:up
create table if not exists metadata (
    id text primary key,
    tool_description text not null,
    table_name text not null,
    columns jsonb not null,
    query_examples jsonb not null,
    created_at timestamp with time zone not null default current_timestamp,
    updated_at timestamp with time zone not null default current_timestamp
);

-- migrate:down
drop table if exists metadata;
