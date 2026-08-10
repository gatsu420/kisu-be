-- migrate:up
create table user_information (
    id uuid primary key default gen_random_uuid(),
    email text unique not null,
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

-- migrate:down
drop table user_information;
