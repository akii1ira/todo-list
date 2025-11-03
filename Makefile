.PHONY: build up down restart clean run

build:
	docker compose build

up:
	docker compose up -d

down:
	docker compose down

restart: down up

clean:
	docker system prune -f

run:
	go run main.go
