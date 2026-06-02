# tashSphere-api

## Project Description

* TaskSphere is a multi-user TODO list management application with an emphasis on privacy by default and the possibility of controlled collaboration

* The system allows each user to create, organize and manage their own tasks completely privately: no other user can see another person's personal tasks unless the owner explicitly decides to share them

* Collaborative tasks can only be seen by those who are part of the group.

## Main features

* Chi framework
* User registration and login (secure authentication) - JWT token and refresh token implemented
* Runtime metrics endpoint for API response times, entity counts, deployment details and architecture notes
* Tasks created could be tagged as private, public or collaborative
* Support for group (collaborative) tasks
* A task can have a owner and multiples assignees (taskEditors)
* All assigned users can view public tasks
* Collaborative tasks can only been seen by memeber of the group
* A Editing a collaborative tasks can be done by the author, the admin and the assigned users
* A collaborative task can be deleted by the author or de admin taks group
* Data is saved in Postgres

## Motivation

After months of studying learning to program in Go, I needed to build a real life project to show my programming skills. It took me a while to figure it what I wanted to build. It must be something could be applied in real solving problem situation.
So I decided to build a **taskSphere** api. Creating the user, tokens, databse was easy. The challenge was to create tasks in a way that could be seen and changed depending of certain permissions.
To update a task is need to check if the logged in user is the author (user_id) of the task or assigned editor (task_editors). In order to do this, I need to compare the user from the JWT token with the task.user_id and the loop in the table of task_editors to check if the user from the token is in the user_id column.

## ⚙️ Installation

Inside a Go module:

```bash
go get github.com/carloscfgos1980/taskSphere-api
```

## 🚀 Quick Start Consumer

```bash
go cmd/*.go
```

## 📖 Usage

### programs needed to run the api

1. postgres
2. goose (migrations)
3. SQLC (generate Go code from SQL queries)
4. pgx driver

### user password

* It must contain at least one capital letter, lowercase letter, one special character and a number

### Recommended fields per task

* Title (required)
* Description
* Creation date
* Deadline/expiration date
* Priority (low, medium, high, urgent)
* Tags (private, , public, collaborative)
* State (pending, in progress, done, cancelled)
* List of assigned users/participants (taskEditors)
* Original creator (user_id)

### tasks

* A task could be personal (private or public) or collaborative.
* Collaborative are the tasks for a group. The parent taks would have empty parent_id and the subtree tasks must has parent_id filled with the refrenced main task (task_id)
* end_time format: 2026-03-22T08:00:00Z
* To view collaborative tasks the parent_id and tag=collaborative most be provided in a query params. The token should match any one in the group
* To view private tasks tag=private most be provided in a query params and the token should match the user
* To view public tasks tag=public most be provided in a query params and mosst be logged in, it does not need to match the user
* Only task_editors assigned by the author of the taks can modified the task
* Only the author of the task can errased

### metrics

* Endpoint: `GET /metrics`
* Purpose: expose runtime and API telemetry in JSON format
* Main fields:
	* `api_response_times_ms`: average and last request duration in milliseconds
	* `requests`: total number of served requests and status class counters (`2xx`, `4xx`, `5xx`)
	* `counts`: current `users` and `tasks` totals from the database
	* `deployment`: bind address, environment (`APP_ENV`), Go version and server start time
	* `architecture`: key implementation decisions (router, auth strategy, database stack, service layering)

Example:

```json
{
	"service": "taskSphere-api",
	"uptime_seconds": 420.51,
	"api_response_times_ms": {
		"average": 6.1,
		"last": 4.8
	},
	"requests": {
		"total": 112,
		"status_2xx": 103,
		"status_4xx": 7,
		"status_5xx": 2
	},
	"counts": {
		"users": 15,
		"tasks": 87
	},
	"deployment": {
		"bind_address": ":8080",
		"app_env": "dev",
		"go_version": "go1.24.0",
		"started_at": "2026-06-02T20:10:00Z"
	},
	"architecture": {
		"router": "chi",
		"auth": "JWT access token + refresh token",
		"database": "PostgreSQL via pgx + sqlc",
		"service_structure": "handler -> service -> database"
	}
}
```

## 🤝 Contributing

### Clone the repo

```bash
git clone -b chi_framework https://github.com/carloscfgos1980/taskSphere-api.git
cd taskSphere-api
```

### Build the compiled binary

```bash
go build
```

### Submit a pull request

If you'd like to contribute, please fork the repository and open a pull request to the `main` branch.
