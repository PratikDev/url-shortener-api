include .env
export

migrate-create:
	sudo docker run --rm \
	-v $(shell pwd)/migrations:/migrations \
	migrate/migrate \
	create \
	-ext sql \
	-dir /migrations \
	-seq $(name)

migrate-up:
	sudo docker run --rm \
	--network url-shortener-api_default \
	-v $(shell pwd)/migrations:/migrations \
	migrate/migrate \
	-source file://migrations \
	-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
	-verbose \
	up

migrate-down:
	sudo docker run --rm \
	--network url-shortener-api_default \
	-v $(shell pwd)/migrations:/migrations \
	migrate/migrate \
	-source file://migrations \
	-database "postgres://$(POSTGRES_USER):$(POSTGRES_PASSWORD)@$(POSTGRES_HOST):$(POSTGRES_PORT)/$(POSTGRES_DB)?sslmode=disable" \
	-verbose \
	down 1

up:
	sudo docker compose up -d --build

down:
	sudo docker compose down

clean:
	sudo docker compose down --remove-orphans -v

clean-hard:
	sudo docker compose down --remove-orphans -v --rmi "all"