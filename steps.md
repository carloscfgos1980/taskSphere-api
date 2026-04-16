# TaskPhere using gin framework

## 1. set up

Get framework for go
go get -u github.com/gin-gonic/gin

Driver for postgres
go get github.com/lib/pq

Package to manage .env
go get github.com/joho/godotenv

* I got an issue with DB_URL coz api config is not in the same directory that .env

Get package for uuid and hash password
go get github.com/google/uuid
go get github.com/alexedwards/argon2id

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

## 2. Create user

1. Create utils/utils.go
This will hold auxiliar functions (HashPassword and IsStrongPassword)

2. user handler /handlers/usr_handler.go
2.1 structs and handler for creating a new user in the system
2.2 UserRequest is the struct for the request body when creating a new user
2.3 CreateUser is the handler for creating a new user in the system
2.3.1 Return a handler function that can be used in the Gin router
2.3.2 Bind the JSON request body to the UserRequest struct
2.3.3 Validate the password strength
2.3.4 Hash the password before storing it in the database
2.3.5 Create the user in the database using the provided configuration and request data
2.3.6 Return the created user as a response, excluding the password
2.3.7 Send the response back to the client with a 200 OK status
