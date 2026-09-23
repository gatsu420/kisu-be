-- migrate:up
alter table tool
add column type text not null default '';

-- migrate:down
alter table tool
drop column type;
