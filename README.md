# TaskPhere using framework

## 1. set up

Get framework for go
go get -u github.com/gin-gonic/gin

Driver for postgres
go get -u github.com/jackc/pgx/v5

PAckage to manage .env
go get github.com/joho/godotenv

* I got an issue with DB_URL coz api config is not in the same directory that .env
