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

## 4. Refresh token

1. function to get token from the request /internal/utils/utils.go

2. Handler to refresh token /internal/handlers/refresh_handler.go
2.1 Define a response struct to return the new access token in the response body
2.2 Return a handler function that can be used in the Gin router
2.3 Extract the refresh token from the Authorization header of the incoming request
2.4 Retrieve the user associated with the provided refresh token from the database
2.5 Generate a new JWT token for the user to authenticate future requests
2.6 Return the new JWT token in the response to the client

3. Register token-related routes /cmd/maing.go

## 5. Revoke refresh token

1. RevokeRefreshTokenHandler is the handler for revoking refresh tokens /internal/handlers/refresh_handler.go
1.1 Return a handler function that can be used in the Gin router
1.2 Extract the refresh token from the Authorization header of the incoming request
1.3 Revoke the provided refresh token in the database to prevent further use
1.4 Return a success message in the response to the client

2. Register token-related routes /cmd/maing.go

## 6. Middleware

/internal/middleware/auth_middleware.go

AuthMiddleware is a Gin middleware function that validates JWT tokens in incoming requests to protect routes that require authentication. It checks for the presence of a valid JWT token in the Authorization header of the request, verifies the token using the secret key from the configuration, and sets the user ID in the Gin context for use in subsequent handlers if the token is valid. If the token is missing or invalid, it returns a 401 Unauthorized response and aborts further processing of the request.

1. Return a handler function that can be used in the Gin router as middleware for routes that require authentication.
2. Extract the Authorization header from the incoming request
3. If there is an error extracting the token (e.g., missing or malformed header), return a 401 Unauthorized response with an appropriate error message and abort the request processing.
4. If the token is valid, set the user ID in the Gin context (e.g., using c.Set("userID", userID)) for use in subsequent handlers that require authentication.
5. Call the next handler in the chain to continue processing the request after successful authentication.

## 7. Create Task router /internal/handlers/task_handler.go

1. Create functios to check request fields (CheckPriority, CheckState and CheckTag) /internal/utils/utils.go
2. Struct of Task that represents a task in the system
3. CreateTaskHandler is the handler for creating a new task in the system
3.1 Define the expected parameters for creating a new task and the response structure
3.2 Return a handler function that can be used in the Gin router
3.3 Extract the user ID from the context (set by the authentication middleware)
3.4 Bind the incoming JSON request to the parameters struct
3.5 Validate the required fields and the values of priority, state, and tag if provided
3.6 Create the task in the database using the provided configuration and parameters
3.7 Prepare the response struct with the created task information
3.8 Return the created task in the response with a 201 Created status

4. Register task-related routes

## 8. Get Task by Id

/internal/handlers/task_handler.go

1. GetTasksByIdHandler is the handler for retrieving a task by its ID
1.1 Return a handler function that can be used in the Gin router
1.2 Extract the task ID from the URL parameters and validate it
1.3 Parse the task ID string into a UUID format
1.4 Retrieve the task from the database using the provided configuration and task ID
1.5 Check if the user making the request is the owner of the task or has access to it
1.6 Prepare the response struct with the retrieved task information
1.7 Return the retrieved task in the response with a 200 OK status

2. Register task-related routes /cmd/main.go

## 9. Get Tasks Handler

1-> GetTasksHandler is the handler for retrieving tasks based on the provided tag and user access
1.1 Return a handler function that can be used in the Gin router
1.2 Extract the tag query parameter and validate it
1.3 Extract the user ID from the context (set by the authentication middleware)

1.4 Retrieve tasks based on the tag and user access level using the provided configuration. Here I used **switch** as condictional
1.4.1 For private tasks, retrieve only the tasks that are owned by the user from the database and return them in the response
1.4.2 Retrieve tasks that are private to the user from the database
1.4.3 If no private tasks are found for the user, return a 404 Not Found response
1.4.4 Prepare the response struct with the retrieved tasks information
1.4.5 Return the retrieved tasks in the response with a 200 OK status

1.5 For collaborative tasks, the user must provide a parent_id query parameter to specify which collaborative tasks to retrieve
1.5.1 Extract the parent_id query parameter and validate it
1.5.2 Parse the parent_id string into a UUID format
1.5.3 Retrieve collaborative tasks that are associated with the specified parent ID from the database
1.5.4 If no collaborative tasks are found for the specified parent ID, return a 404 Not Found response
1.5.5 Prepare the response struct with the retrieved collaborative tasks information, including user details and task editors
1.5.6 Check if the user making the request is the owner of any of the collaborative tasks or has access to them
1.5.7 If the user does not have access to any of the collaborative tasks, return a 403 Forbidden response
1.5.8 Return the retrieved collaborative tasks in the response with a 200 OK status

1.6 For public tasks, retrieve all tasks that are tagged as public from the database and return them in the response
1.6.1 Retrieve public tasks from the database
1.6.2 If no public tasks are found, return a 404 Not Found response
1.6.3 Prepare the response struct with the retrieved public tasks information, including user details
1.6.4 Return the retrieved public tasks in the response with a 200 OK status

1.7 If the tag value is not valid, return a 400 Bad Request response

2-> Register task-related routes

## 10. Get collaborative tasks parent ID

1. sql query to get a list of collaborative task parents
1.1 Generate go code with sqlc command

2. GetCollaborativeTasksHandler is the handler for retrieving collaborative tasks that are associated with a specified parent ID
2.1 Return a handler function that can be used in the Gin router
2.2 Extract the user ID from the context (set by the authentication middleware)
2.3 Check if the user making the request is valid and exists in the database
2.4 Retrieve collaborative tasks that are associated with the specified parent ID from the database
2.5 If no collaborative tasks are found for the specified parent ID, return a 404 Not Found response
2.6 Prepare the response struct with the retrieved collaborative tasks information, including user details
2.7 loop through the retrieved collaborative tasks and then prepare the response struct with the task information and user details
2.8 Return the retrieved collaborative tasks in the response with a 200 OK status

3. Register task-related routes

## 11. Update task
