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
	
.PHONY: grpc
grpc:
	go run ./cmd/gobank/main.go -grpc

.PHONY: mock
mock:
	mockgen -package mockdb -destination internal/db/mock/store.go github.com/lucasHSantiago/gobank/internal/db/sqlc Store

.PHONY: proto
proto:
	rm -f proto/gen/*.go
	rm -f docs/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=proto/gen --go_opt=paths=source_relative \
    --go-grpc_out=proto/gen --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=proto/gen --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=docs/swagger --openapiv2_opt=allow_merge=true,merge_file_name=gobank \
	proto/*.proto
	statik -src=./docs/swagger -dest=./docs

.PHONY: grpcui
grpcui:
	grpcui -plaintext localhost:9090

.PHONY: redis
redis:
	docker run --name redis -p 6379:6379 -d redis:7.4.2-alpine
