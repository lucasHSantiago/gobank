.PHONY: postgres
postgres:
	docker run --name postgres --network bank-network -p 5432:5432 -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=admin -d postgres:alpine

.PHONY: docker
docker:
	docker run --name gobank --network bank-network -p 8080:8080 -e GIN_MODE=release -e DB_SOURCE="postgresql://postgres:admin@postgres:5432/gobank?sslmode=disable" gobank:lates

.PHONY: createdb
createdb:
	docker exec -it postgres createdb -U postgres --username=postgres --owner=postgres gobank

.PHONY: dropdb
dropdb:
	docker exec -it postgres dropdb -U postgres gobank

.PHONY: migrateup
migrateup:
	migrate -path internal/db/migrations -database "postgresql://postgres:admin@localhost:5432/gobank?sslmode=disable" -verbose up $(or $(n))

.PHONY: migratedown
migratedown:
	migrate -path internal/db/migrations -database "postgresql://postgres:admin@localhost:5432/gobank?sslmode=disable" -verbose down $(or $(n))

.PHONY: sqlc
sqlc:
	sqlc generate

.PHONY: test
test:
	go test -v -cover ./...

.PHONY: server
server:
	go run ./cmd/gobank/main.go

.PHONY: mock
mock:
	mockgen -package mockdb -destination internal/db/mock/store.go github.com/lucasHSantiago/gobank/internal/db/sqlc Store
