-- migrate:up
alter table tool
add column param_name text not null default '',
add column param_type text not null default '',
add column param_description text not null default '';

-- migrate:down
alter table tool
drop column param_name,
drop column param_type,
drop column param_description;
