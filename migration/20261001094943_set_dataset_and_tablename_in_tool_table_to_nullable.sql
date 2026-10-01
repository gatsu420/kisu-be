-- migrate:up
alter table tool
alter column dataset drop not null,
alter column table_name drop not null;

-- migrate:down
alter table tool
add column dataset text not null default '',
add column table_name text not null default '',
