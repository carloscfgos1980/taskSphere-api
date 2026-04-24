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

package to create jwt
go get github.com/golang-jwt/jwt/v5

package to encrypt password
go get github.com/alexedwards/argon2id

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

## 2. Create user

1. Create package that holds utils functions (IsStrongPassword and IsValidEmail)
2. Create package to handle to read the request and write the response in json /json/json.go
3. Define type (structs) needed to register a new user /internal/users/types.go

4. Service set up /internal/users/service.go
4.1 Service defines the interface for the users service
4.2 svc defines the struct for the users service

5. CreateUser creates a new user in the database /internal/users/service.go
5.1 start a transaction
5.2 create a new Queries instance with the transaction
5.3 create the user
5.4 commit the transaction
5.5 return the created user

6. Handler set up /internal/users/handler.go
6.1 handler is the HTTP handler for users endpoints
6.2 NewHandler creates a new handler for users endpoints

7. CreateUser handles the HTTP request for creating a new user /internal/users/handler.go
7.1 Parse the JSON request body into a UserRequest struct
7.2 Check if any field is empty
7.3 Validate email format
7.4 Validate the password strength
7.5 Hash the password before storing it in the database
7.6 Update the user request with the hashed password
7.7 Call the service to create the user
7.8 Create a response struct to send back to the client, excluding the password
7.9 Write the response as JSON with a 201 Created status code

8. users endpoints. Create the user service and handler /cmd/api.go
9. set up the users routes /cmd/api.go
