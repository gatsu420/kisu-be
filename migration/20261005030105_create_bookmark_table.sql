-- migrate:up
create table bookmark (
    id uuid primary key,
    user_id uuid not null references user_information(id),
    name text not null,
    param_name text not null,
    param_value text not null,
    query text not null,
    hashed_tool text not null,
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

-- migrate:down
drop table bookmark;
