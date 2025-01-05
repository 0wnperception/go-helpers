mkdir -p    cmd \
            internal/models \
            internal/ports/rest \
            internal/ports/grpc \
            internal/adapters/repo \
            internal/services \
            internal/app \
            migrations \
            scripts \
            configs \

echo "package main
func main(){

}" > cmd/main.go 

echo "dev:
    dialect: postgres
    datasource: sslmode=disable host=\${DB_HOST} dbname=\${DB_NAME} user=\${DB_USER} password=\${DB_PASSWORD} port=\${DB_PORT}
    dir: migrations
    table: migrations" > migrations/dbconfig.yaml

echo "general:
    name: ${PROJECT_NAME}
db:
    host: localhost
    name: postgres
    user: postgres
    password: postgres
    port: 5432" > configs/config.yaml

echo "package configs
import (
	\"fmt\"
	\"github.com/kelseyhightower/envconfig\"
)
type Config struct {
	General struct {
		Name string \`yaml:\"name\" default:\"${PROJECT_NAME}\"\`
	} \`yaml:\"general\"\`
	DB struct {
		Host     string \`yaml:\"host\" env:\"PG_HOST\"\`
		Name     string \`yaml:\"name\" env:\"PG_DBNAME\"\`
		User     string \`yaml:\"user\" env:\"PG_USER\"\`
		Password string \`yaml:\"password\" env:\"PG_PASSWORD\"\`
		Port     string \`yaml:\"port\" env:\"PG_PORT\" default:\"5432\"\`
	} \`yaml:\"db\"\`
}" > configs/config.go


