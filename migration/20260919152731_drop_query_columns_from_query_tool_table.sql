-- migrate:up
alter table query_tool
drop column query,
drop column tool_description;

-- migrate:down
alter table query_tool
add column query text not null default '',
add column tool_description text not null default '',
