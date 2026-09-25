-- migrate:up
alter table tool
drop column param_type,
drop column param_description;

-- migrate:down
alter table tool
add column param_type text not null default '',
add column param_description text not null default '';
