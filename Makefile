dev:
	air -- -env=./.env
run/server:
	cd server && go run .	
run/seaweed:
	weed server -dir="./tmp/seaweed" -s3 -s3.config=seaweed-config.json
gen/proto:
	protoc proto/*.proto --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative  --ts_out=. --ts_opt=unary_rpc_promise=true
migrate/dev:
	atlas schema apply --url "sqlite://app.db" --dev-url "sqlite://dev.db" --to "file://sql/schema.sql" --auto-approve