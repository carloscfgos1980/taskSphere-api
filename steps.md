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

## 3. Login

1. Config JWT Secret /cmd/main.go
1.1 Get the JWT secret from environment variables
1.2 Add JWT to cfg (config)

2. Add services /internal/users/service.go
2.1 GetUserByEmail gets a user from the database by email
2.2 CreateRefreshToken creates a new refresh token for a user in the database

3. Add JWT field to the handler returned from NewHandler /internal/users/handler

4. LoginUser handles the HTTP request for logging in a user
4.1 Parse the JSON request body into a UserRequest struct
4.2 Check if email and password are provided
4.3 Get the user by email from the database
4.4 Check if the provided password matches the stored hashed password
4.5 Generate a JWT token for the authenticated user
4.6 Generate a refresh token and store it in the database
4.7 Create a response struct to send back to the client with the access token
4.8 Write the response as JSON with a 200 OK status code

5. Add JWT as second parameter to the function NewHandler:
 userHandler := users.NewHandler(userService, app.config.JWTSecret)

Note: This was a bit challenging coz I am used to access config directly and in this case it has to be done thru the handler
6. set up the users routes (login)

## 4. Refresh token

1. set up for service /internal/refresh/service.go
1.1 Service defines the interface for the users service
1.2 svc defines the struct for the users service
1.3 NewService creates a new service for the users package

2. Defive service for refresh handler
2.1 GetUserFromRefreshToken retrieves the user associated with the given refresh token from the database /internal/refresh/service.go
2.2 add GetUserFromRefreshToke to Service interface

3. set up for handler
3.1 handler is the HTTP handler for users endpoints
3.2 NewHandler creates a new handler for users endpoints

4. RefreshToken handles the HTTP request for refreshing a user's access token
4.1 Get the refresh token from the Authorization header
4.2 Get the user associated with the refresh token from the database
4.3 Generate a new JWT token for the user to authenticate future requests
4.4 Create a response struct to send back to the client with the new access token and refresh token
4.5 Write the response as JSON with a 200 OK status code

5. refresh token endpoint /cmd/api.go

## 5. Revoke refresh token

1. Service for revoke the refresh token /internal/refresh/service.go
1.1 RevokeRefreshToken method revokes a refresh token in the database, preventing it from being used to generate new access tokens
1.2 Add RevokeRefreshToken method to Service interface

2. RevokeRefreshToken handles the HTTP request for revoking a user's refresh token /internal/refresh/handler.go
2.1 Get the refresh token from the Authorization header
2.2 Revoke the refresh token in the database
2.3 Write a success response with a 200 OK status code

3. refresh token endpoint /cmd/api.go

## 6. Middleware

1. HTTP middleware setting a value on the request context /internal/authmiddleware/auth_middleware.go
1.1 Return a new http.HandlerFunc that wraps the original handler and adds the authentication logic
1.2 Extract the token from the Authorization header
1.3 Validate the token and extract the user ID
1.4 Create a new context with the user ID value
1.5 Call the next handler with the new context

2. protected routes

```go
 r.Route("/api", func(r chi.Router) {
  // Add authentication middleware here if available
  r.Use(func(next http.Handler) http.Handler {
   return authmiddleware.AuthMiddleware(next, app.config.JWTSecret)
  })
 })
```

## 7. Create a task

1. Task struct represents a task in the system /internal/tasks/types.go
2. Set up srvice /internal/tasks/service.go
2.1 Service defines the interface for the users service
2.2 svc defines the struct for the users service
2.3 NewService creates a new service for the users package

3. GetUserByID method of svc retrieves a user by their ID from the database
4. CreateTask method of svc creates a new task in the database associated with the given user ID
5. Add "GetUserByID" and "CreateTask" to service interface

6. Set up for the handler /internal/tasks/handler.go
6.1 handler is the HTTP handler for users endpoints
6.2 NewHandler creates a new handler for users endpoints

7. CreateTask handles the creation of a new task for a user
7.1 Define the expected parameters for creating a new task and the response structure
7.2 Get the user ID from the request context (set by the authentication middleware)
7.3 Check if the user ID is present in the context
7.4 Assert the user ID value to a UUID type
7.5 Convert the user ID to a string for database queries
7.6 Check if the user exists in the database
7.7 Parse the JSON request body into a TaskRequest struct
7.8 Check if any field is empty
7.9 Validate the priority, state, and tag values
7.10 Create a Task struct with the parsed data and the user ID
7.11 Call the service to create the task in the database
7.12 Convert []pgtype.UUID to []uuid.UUID
7.13 Create a response struct to send back to the client
7.14 Write the response as JSON

8. create the task service and handler cmd/api.go

## 8. Get a task

1. GetTaskByID method of svc struct retrieves a task by its ID from the database /internal/tasks/service.go
2. Add GetTaskByID to service interface

3. GetTaskByID handles the retrieval of a task by its ID
3.1 Get the task ID from the URL parameters
3.2 Call the service to get the task from the database
3.3 Get the user ID from the request context (set by the authentication middleware)
3.4 Check if the user ID is present in the context
3.5 Assert the user ID value to a UUID type
3.6 Convert the user ID to a string for database queries
3.7 Check if the user has access to the task (is the owner or an editor)
3.8 Check if the user is an editor of the task
3.9 If the user is not the owner and not an editor, return a forbidden error
3.10 Convert []pgtype.UUID to []uuid.UUID
3.11 Create a response struct to send back to the client
3.12 Write the response as JSON

4. create the task service and handler /cmd/api.go

## 9. Get taks

1. GetTasksByUserID method of svc retrieves all tasks associated with a given user ID from the database internal/tasks/service.go
2. GetCollaborativeTasksByParentID retrieves all collaborative tasks associated with a given parent task ID or task ID if that user is the owner of the parent ID from the database
3. GetPublicTasks retrieves all public tasks from the database
4. Add GetTasksByUserID, GetCollaborativeTasksByParentID and GetPublicTasks to service interface

5. GetTasks handles the retrieval of tasks based on the tag query parameter and user access level /internal/tasks/handler.go
5.1 Extract the tag query parameter and validate it
5.2 Extract the user ID from the context (set by the authentication middleware)
5.3 Assert the user ID value to a UUID type
5.4 Convert the user ID to a string for database queries

5.5 switch tag. Handle the retrieval of tasks based on the tag value and user access level
5.5.1 If the tag is "private", retrieve tasks that are private to the user from the database
5.5.1.1 Retrieve tasks that are private to the user from the database
5.5.1.2 If no private tasks are found for the user, return a 404 Not Found response
5.5.1.3 Prepare the response struct with the retrieved tasks information
5.5.1.4 Convert []pgtype.UUID to []uuid.UUID
5.5.1.5 Create a response struct for each task with the retrieved information, including user details and task editors
5.5.1.6 Write the retrieved tasks as JSON response

5.5.2 If the tag is "collaborative", retrieve tasks that are collaborative and associated with the specified parent ID from the database
5.5.2.1 Extract the parent_id query parameter and validate it
5.5.2.2 Retrieve collaborative tasks that are associated with the specified parent ID from the database
5.5.2.3 If no collaborative tasks are found for the specified parent ID, return a 404 Not Found response
5.5.2.4 Prepare the response struct with the retrieved collaborative tasks information, including user details and task editors
5.5.2.5 Check if the user making the request is the owner of any of the collaborative tasks or has access to them
5.5.2.6 If the user does not have access to any of the collaborative tasks, return a forbidden error
5.5.2.7 Write the retrieved collaborative tasks as JSON response

5.5.3 If the tag is "public", retrieve tasks that are public from the database
5.5.3.1 Retrieve tasks that are public from the database
5.5.3.2 If no public tasks are found, return a 404 Not Found response
5.5.3.3 Prepare the response struct with the retrieved public tasks information, including user details and task editors
5.5.3.4 Convert the retrieved public tasks to the response struct format, including user details and task editors
5.5.3.5 Write the retrieved public tasks as JSON response

5.5.4 If the tag value is invalid, return a bad request error

1. create the task service and handler /cmd/api.go

## 10. Get collaborative tasks parents

1. Move the struct taskResponse that represents the response format for a task, including the username and email of the user who created the task to types to avoid repeating code /internal/tasks/types.go
2. GetParentTasks  method of svc retrieves all parent tasks from the database /internal/tasks/service.go
3. Add GetParentTasks method to service interface

4. GetParentsCollaborativeTasks handles the retrieval of parent tasks that are collaborative and associated with the user making the request from the database /internal/tasks/handler.go
4.1 Extract the user ID from the context (set by the authentication middleware)
4.2 Assert the user ID value to a UUID type
4.3 Convert the user ID to a string for database queries
4.4 Check if the user exists in the database
4.5 Retrieve parent tasks that are collaborative
4.6 Write the response as JSON

5. set up the tasks routes /cmd/api.go

```go
  r.Get("/tasks/collaborative", taskHandler.GetParentsCollaborativeTasks)
```

## 11. Update a task

1. UpdateTask method of svc updates a task in the database /internal/tasks/service.go
2. Add UpdateTask method to service interface

3. UpdateTask handles the updating of a task by its ID, allowing only the owner or editors of the task to perform the update /internal/tasks/handler.go
3.1 Define a struct to hold the parameters for updating the task
3.2 Get the task ID from the URL parameters
3.3 Validate the task ID
3.4 Get the user ID from the request context (set by the authentication middleware)
3.5 Check if the user ID is present in the context
3.6 Assert the user ID value to a UUID type
3.7 Convert the user ID to a string for database queries
3.8 Get the task from the database to check if it exists and to verify the user's access level (owner or editor) for authorization to update the task
3.9 Check if the user making the request is the owner of the task or has access to it (either as an editor or as a collaborator) to determine if they are authorized to update the task
3.10 Parse the JSON request body into the parameters struct to get the fields that need to be updated for the task
3.11 Set missing fields to their current values in the database to ensure that only the provided fields are updated while the others remain unchanged
3.12 Create a Task with the updated fields and the user ID to pass to the service for updating the task in the database, ensuring that only the provided fields are updated while the others remain unchanged
3.13 Call the service to update the task in the database with the provided fields, ensuring that only the provided fields are updated while the others remain unchanged
3.14 Create a response struct to send back to the client with the updated task information
3.15 Write the response as JSON

4. set up the tasks routes /cmd/api.go

```go
  r.Put("/tasks/{taskID}", taskHandler.UpdateTask)
```

## 12. Delete a task

