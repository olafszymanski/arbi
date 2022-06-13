include .env

up:
	docker-compose up --build

down:
	docker-compose down

db_up:
	docker-compose up -d
	docker run -v $(MIGRATIONS_PATH):/migrations --network host migrate/migrate -path=/migrations -database $(DB_DRIVER)://$(DB_USER):$(DB_PASSWORD)@localhost:5432/$(DB_NAME)?sslmode=disable up
	docker-compose down

db_down:
	docker-compose up -d
	docker run -v $(MIGRATIONS_PATH):/migrations --network host migrate/migrate -path=/migrations -database $(DB_DRIVER)://$(DB_USER):$(DB_PASSWORD)@localhost:5432/$(DB_NAME)?sslmode=disable down
	docker-compose down