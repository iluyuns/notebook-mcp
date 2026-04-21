# 迁移目录（相对于本 Makefile 所在目录）
MIGRATE_DIR := migrations

# 与 docker-compose、.env.example 一致；可被环境变量或命令行覆盖
POSTGRES_HOST ?= 192.168.0.165
POSTGRES_PORT ?= 5432
POSTGRES_USER ?= notebook
POSTGRES_PASSWORD ?= notebook
POSTGRES_DB ?= notebook

LOCAL_PSQL_DSN := postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable

# 优先级：DATABASE_URL > POSTGRES_DSN > 上面本地默认
DBURL := $(if $(DATABASE_URL),$(DATABASE_URL),$(if $(POSTGRES_DSN),$(POSTGRES_DSN),$(LOCAL_PSQL_DSN)))

.PHONY: mu md mda mv mf mc migrate-up migrate-down migrate-down-all migrate-version migrate-force migrate-create

mu: migrate-up
md: migrate-down
mda: migrate-down-all
mv: migrate-version
mf: migrate-force
mc: migrate-create

migrate-up:
	migrate -path $(MIGRATE_DIR) -database "$(DBURL)" up

migrate-down:
	migrate -path $(MIGRATE_DIR) -database "$(DBURL)" down 1

migrate-down-all:
	migrate -path $(MIGRATE_DIR) -database "$(DBURL)" down -all

migrate-version:
	migrate -path $(MIGRATE_DIR) -database "$(DBURL)" version

migrate-force:
	@test -n "$(VERSION)" || (echo "usage: make mf VERSION=1" >&2 && false)
	migrate -path $(MIGRATE_DIR) -database "$(DBURL)" force $(VERSION)

migrate-create:
	@test -n "$(NAME)" || (echo "usage: make mc NAME=add_xxx" >&2 && false)
	migrate create -seq -ext sql -dir $(MIGRATE_DIR) -digits 6 $(NAME)
