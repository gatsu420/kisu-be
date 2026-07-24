.PHONY: migration-new
migration-new:
	dbmate \
		--migrations-dir "./migration" \
		new $(NAME)

.PHONY: migration-up
migration-up:
	dbmate \
		--env POSTGRES_DSN \
		--migrations-dir "./migration" \
		--schema-file "./migration/schema.sql" \
		up

.PHONY: migration-down
migration-down:
	dbmate \
		--env POSTGRES_DSN \
		--migrations-dir "./migration" \
		--schema-file "./migration/schema.sql" \
		down
