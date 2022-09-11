postgres:
	docker run --name live-stream-postgres12 -p 5432:5432 -v /home/donjuan/dockerpostgresql:/var/lib/postgresql/data -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine
createdb:
	docker exec -it live-stream-postgres12 createdb --username=root --owner=root live-stream
create migration:
	migrate create -ext sql -dir app/db/migration -seq init_chema
migrateup:
	migrate -path app/db/migration -database "postgresql://root:secret@localhost:5432/live-stream?sslmode=disable" -verbose up
migratedown:
	migrate -path app/db/migration -database "postgresql://root:secret@localhost:5432/live-stream?sslmode=disable" -verbose down
migrateup1:
	migrate -path app/db/migration -database "postgresql://root:secret@localhost:5432/live-stream?sslmode=disable" -verbose up 1
migratedown1:
	migrate -path app/db/migration -database "postgresql://root:secret@localhost:5432/live-stream?sslmode=disable" -verbose down 1
sqlc:
	sqlc generate -f ./app/sqlc.yaml