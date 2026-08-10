-- migrate:up
create unique index user_id_index on user_token (user_id);

-- migrate:down
drop index user_id_index;
