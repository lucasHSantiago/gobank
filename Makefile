.PHONY: postgres
postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=admin -d postgres:alpine

.PHONY: createdb
createdb:
	docker exec -it postgres createdb -U postgres --username=postgres --owner=postgres gobank

.PHONY: dropdb
dropdb:
	docker exec -it postgres dropdb -U postgres gobank

.PHONY: migrateup
migrateup:
	migrate -path db/migrations -database "postgresql://postgres:admin@localhost:5432/gobank?sslmode=disable" -verbose up

.PHONY: migratedown
migratedown:
	migrate -path db/migrations -database "postgresql://postgres:admin@localhost:5432/gobank?sslmode=disable" -verbose down

.PHONY: sqlc
sqlc:
	sqlc generate

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: server
server:
	go run main.go
