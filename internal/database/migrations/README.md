# Migrations

## install

go install -tags 'pgx5' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

## migrate manually

migrate -database pgx5://postgres:example@localhost:5432/whodidthechores?sslmode=disable -path ./migrations up
migrate -database pgx5://postgres:example@localhost:5432/whodidthechores?sslmode=disable -path ./migrations down

## create migration

migrate create -dir ./migrations -ext sql -seq create_chores_table