.PHONY: build run test migrate rollback

build:
	docker-compose build

run:
	docker-compose up

test:
	go test ./...

migrate:
	go run cmd/migrate/main.go

migrate-admin:
	go run cmd/migrate/main.go -use-admin

rollback-admin:
	@echo "Rolling back the last migration with admin credentials..."
	@read -p "Are you sure? (y/n) " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		go run cmd/migrate/main.go -rollback -use-admin; \
	fi

rollback:
	@echo "Rolling back the last migration..."
	@read -p "Are you sure? (y/n) " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		go run cmd/migrate/main.go -rollback; \
	fi

worker:
	go run cmd/worker/main.go

api:
	go run cmd/api/main.go

compose-migrate:
	docker-compose run --rm migrate

compose-rollback:
	@echo "Rolling back the last migration..."
	@read -p "Are you sure? (y/n) " -n 1 -r; \
	echo; \
	if [[ $$REPLY =~ ^[Yy]$$ ]]; then \
		docker-compose run --rm migrate ./migrate -rollback; \
	fi