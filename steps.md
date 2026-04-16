# TaskPhere using gin framework

## 1. set up

Get framework for go
go get -u github.com/gin-gonic/gin

Driver for postgres
go get -u github.com/jackc/pgx/v5

Package to manage .env
go get github.com/joho/godotenv

* I got an issue with DB_URL coz api config is not in the same directory that .env

1. Create config.go
2. Create cmd/main.go
2.1 create a context
2.2 Load configuration from environment variables
2.3 Connect to the database using cfg.DB_URL
2.4 Initialize the Gin router
2.5 Set trusted proxies to nil to avoid warnings in Gin 1.7+
2.6 Define a simple health check route
2.7 Start the server on the specified port

3. From taskSphere v2, add sql directory and sqlc.yaml. Then generate go cofe for queries run in the terminal:

```bash
sqlc generate
```
