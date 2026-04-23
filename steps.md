# Steps

## 1. Set up

get framework package:
go get -u github.com/go-chi/chi/v5

get .env package
go get github.com/joho/godotenv

get package driver for database
github.com/jackc/pgx/v5

get uuid package
go get github.com/google/uuid

command to run the app. I got an issue while running just main.go
go run ./cmd/*.go

1. Create api structure to handle the route from chi
1.1 application is the main application struct that holds the configuration and database connection
1.2 config holds the configuration for the application
1.3 dbConfig holds the database configuration for the application

1.4 mount sets up the routes and middleware for the application
1.4.1 create a new router
1.4.2 set up middleware
1.4.3 Set a timeout value on the request context (ctx), that will signal through ctx.Done() that the request has timed out and further processing should be stopped.
1.4.4 health check endpoint
1.4.5 return the router

1.5 run starts the HTTP server
1.5.1 create the HTTP server
1.5.2 start the server
2. Set up to read goose string / internal/env/env.go
3. Set up the server
3.1 Load environment variables from .env file
3.2 Get the port from environment variables, default to 8080 if not set
3.3 create a context
3.4 load env variables
3.5 initialize logger
3.6 database connection
3.7 create the application
3.8 run the application
4. Generate  Go code from sql

```bash
sqlc generate
```

Here I had an issue coz in the sql.yaml file I didn't include the driver (pgx/v5)
