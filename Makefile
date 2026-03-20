# note: call scripts from /scripts

migrate:
	export $(shell cat .env) && \
		migrate -path internal/migrations -database postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable up

migrate-down:
	export $(shell cat .env) && \
		migrate -path internal/migrations -database postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable down $(NUM)

migrate-force:
	export $(shell cat .env) && \
		migrate -path internal/migrations -database postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@$$POSTGRES_HOST:$$POSTGRES_PORT/$$POSTGRES_DB?sslmode=disable force $(VER)
