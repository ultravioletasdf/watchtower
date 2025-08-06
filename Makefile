dev:
	air -- -env=./.env
dev/htmx:
	cd clients/htmx && make dev
dev/video_analyser:
	cd queues/video_analyser && air
dev/transcoder:
	cd queues/transcoder && air
dev/genvtt:
	cd queues/genvtt && air
run/server:
	cd server && go run .
run/web:
	cd clients/web && bun dev
run/seaweed:
	weed server -dir="./tmp/seaweed" -s3 -s3.config=seaweed-config.json
gen/proto:
	protoc proto/*.proto --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative  --ts_out=. --ts_opt=unary_rpc_promise=true
gen/vips:
	vipsgen -out ./vips
migrate/dev:
	atlas schema apply --url "sqlite://app.db" --dev-url "sqlite://dev.db" --to "file://sql/schema.sql" --auto-approve
migrate/devpg:
	atlas schema apply --url "postgres://postgres:dev@localhost:5432/?search_path=public&sslmode=disable" --dev-url "docker://postgres/15/dev?search_path=public" --to "file://sql/schema.sql" --auto-approve
