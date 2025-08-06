minio: minio server ./tmp/minio
rabbitmq: wait-for-it -t 0 localhost:9000 -- docker compose up

queue_video_analsyer: wait-for-it -t 0 localhost:5672 -- make dev/video_analyser
queue_transcoder: wait-for-it -t 0 localhost:5672 -- make dev/transcoder
thumbnail_generator: wait-for-it -t 0 localhost:5672 -- make dev/genvtt

server: wait-for-it -t 0 localhost:5672 -- wait-for-it -t 0 localhost:5432 -- make dev
htmx: wait-for-it -t 0 localhost:50051 -- make dev/htmx
