# Developing

## Dependencies

```sh
go install github.com/air-verse/air@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
curl -sSf https://atlasgo.sh | sh
bun install -g protoc-gen-ts
go install github.com/cshum/vipsgen/cmd/vipsgen@latest
go install github.com/mattn/goreman@latest
```

### S3

For local development you can use MinIO + MinIO Client: [Download](https://min.io/open-source/download?platform=linux])

Once minio is running, you can create an access key with:
```sh
mcli alias set 'dev' 'http://127.0.0.1:9000' 'minioadmin' 'minioadmin'
mcli admin accesskey create dev
```

## Running

### Environment

Copy `.env.example` to `.env` and update the values.
If you have an nvidia gpu, set `TRANSCODE_NVIDIA=true` to make processing videos faster

After that, you can start all needed tools with:
```sh
# Starts minio, rabbitmq, video queues and the server/webclient
goreman start
```
