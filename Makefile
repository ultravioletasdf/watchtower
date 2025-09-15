gen/proto:
	protoc -I proto proto/*.proto --go_out=internal/generated/proto --go_opt=paths=source_relative --go-grpc_out=internal/generated/proto --go-grpc_opt=paths=source_relative
gen/vips:
	vipsgen -out ./internal/generated/vips
migrate/dev:
	atlas schema apply --url "sqlite://app.db" --dev-url "sqlite://dev.db" --to "file://sql/schema.sql" --auto-approve
migrate/devpg:
	atlas schema apply --url "postgres://postgres:dev@localhost:5432/?search_path=public&sslmode=disable" --dev-url "docker://postgres/15/dev?search_path=public" --to "file://sql/schema.sql" --auto-approve
