# tashSphere

## Project Description

* TaskSphere is a multi-user TODO list management application and the possibility of controlled collaboration.
* The system allows each user to create, organize and manage their own tasks completely privately: no other user can see another person's personal tasks

## Main features

* User registration and login (secure authentication) - JWT token and refresh token implemented
* Tasks created could be private or collaborative. Private can only been seen by the author
* Support for group/collaborative tasks:
* A task can have multiple owners/assignees (taskEditors)
* All assigned users can view and (depending on permissions) modify the task
* Editing (modify title, description, status, date, etc.)
* Data is saved in postgres

## programs needed to work with the api

1. Install postgres
2. Install goose (migrations)
3. Install SQLC (generate Go code from SQL queries)

## Typical task states

* Pending
* In progress
* Waiting
* Done
* Canceled

## Recommended fields per task

* Title (required)
* Description
* Creation date
* Deadline/expiration date
* Priority (low, medium, high, urgent)
* Tags (private, collaborative)
* State
* List of assigned users/participants
* Original creator
* Subtasks / checklist

## ⚙️ Installation

Inside a Go module:

```bash
go get github.com/carloscfgos1980/taskSphere-api
```

## 🚀 Quick Start Consumer

```bash
go run .
```
