DB_USERNAME="postgres"
DB_PASSWORD="example"
DB_HOST="localhost"
DB_PORT="5431"
DB_NAME="postgres"

build:
	go build cmd/main.go

run:
	DB_USERNAME=${DB_USERNAME} DB_PASSWORD=${DB_PASSWORD} DB_HOST=${DB_HOST} DB_PORT=${DB_PORT} DB_NAME=${DB_NAME} go run cmd/main.go 

