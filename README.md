# tashSphere-api

## Project Description

* TaskSphere is a multi-user TODO list management application with an emphasis on privacy by default and the possibility of controlled collaboration

* The system allows each user to create, organize and manage their own tasks completely privately: no other user can see another person's personal tasks unless the owner explicitly decides to share them

* Collaborative tasks can only be seen by those who are part of the group.

## Main features

* User registration and login (secure authentication) - JWT token and refresh token implemented
* Tasks created could be tagged as private, publiic or collaborative
* Support for group (collaborative) tasks
* A task can have a owner and multiples assignees (taskEditors)
* All assigned users can view and (depending on permissions) modify the task
* Editing (modify title, description, status, date, etc.)
* Data is saved in postgres

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
go run .
```

## 📖 Usage

### programs needed to run the api

1. postgres
2. goose (migrations)
3. SQLC (generate Go code from SQL queries)

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

* Task could be personal (private or public) or collaborative.
* Collaborative are the tasks for a group. The parent taks would have empty parent_id and the subtree tasks must has parent_id filled with the refrenced main task (task_id)
* end_time format: 2026-03-22T08:00:00Z
* To view collaborative tasks the parent_id must be provided as URL path value and be logged in. Only users of the groud can see the list of tasks
* Only task_editors assigned by the author of the taks can modified the task
* Only the author of the task can errased

## 🤝 Contributing

### Clone the repo

```bash
git clone github.com/carloscfgos1980/taskSphere-api
cd taskSphere-api
```

### Build the compiled binary

```bash
go build
```

### Submit a pull request

If you'd like to contribute, please fork the repository and open a pull request to the `main` branch.
