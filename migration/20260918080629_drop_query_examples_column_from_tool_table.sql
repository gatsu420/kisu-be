-- migrate:up
alter table tool
drop column query_examples;

-- migrate:down
alter table tool
add column query_examples text not null default '';
