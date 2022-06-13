include .env

run:
	docker-compose up --build

stop:
	docker-compose down

up:
	docker run -v $(MIGRATIONS_PATH):/migrations --network host migrate/migrate -path=/migrations -database $(DB_DRIVER)://$(DB_USER):$(DB_PASSWORD)@localhost:5432/$(DB_NAME)?sslmode=disable up

down:
	docker run -v $(MIGRATIONS_PATH):/migrations --network host migrate/migrate -path=/migrations -database $(DB_DRIVER)://$(DB_USER):$(DB_PASSWORD)@localhost:5432/$(DB_NAME)?sslmode=disable down