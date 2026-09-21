-- migrate:up
alter table tool
add column project text not null default '',
add column dataset text not null default '';

-- migrate:down
alter table tool
drop column project,
drop column dataset;
