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
This will hold auxiliar functions (HashPassword, IsValidEmail and IsStrongPassword)

2. Create user handler /internal/handlers/usr_handler.go
2.1 structs and handler for creating a new user in the system
2.2 UserRequest is the struct for the request body when creating a new user
2.3 CreateUser is the handler for creating a new user in the system
2.3.1 Return a handler function that can be used in the Gin router
2.3.2 Bind the JSON request body to the UserRequest struct
2.3.3 Validate email format
2.3.4Validate the password strength
2.3.5 Hash the password before storing it in the database
2.3.6 Create the user in the database using the provided configuration and request data
2.3.7 Return the created user as a response, excluding the password
2.3.8 Send the response back to the client with a 200 OK status

3. Register user-related routes

## 3. Login User

1. Create utils/utils.go
This will hold auxiliar functions (ValidateJWT and MakeRefreshToken)

2. Login User /internal/handlers/user_handler.go
2.1 LoginRequest is the struct for the request body when logging in a user
2.2 LoginUserHandler is the handler for logging in a user and generating a JWT token and refresh token
2.2.1 Define a response struct that includes the user information and the generated tokens
2.2.2 Return a handler function that can be used in the Gin router
2.2.3 Bind the JSON request body to the LoginRequest struct
2.2.4 Validate email format
2.2.5 Retrieve the user from the database using the provided email
2.2.6 Check if the provided password matches the stored hashed password
2.2.7 Generate a JWT token for the authenticated user
2.2.8 Generate a refresh token and store it in the database
2.2.9 Return the user information along with the generated tokens in the response
2.2.10 Send the response back to the client with a 200 OK status

3. Register user-related routes
