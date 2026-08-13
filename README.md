# Task Manager - Backend

A RESTful backend for a full-stack Task Management application built with Go, Gin, GORM, and MySQL.

The backend provides user authentication, JWT authorization, task management, admin functionality, and password recovery through email.

## Tech Stack

- Go
- Gin
- GORM
- MySQL
- JWT Authentication
- bcrypt
- Gmail SMTP
- godotenv

## Features

### Authentication

- User registration
- User login
- Password hashing using bcrypt
- JWT-based authentication
- Role-based authorization
- Protected routes
- Admin-only routes

### Task Management

- Create tasks
- View tasks
- View individual tasks
- Update tasks
- Delete tasks
- Task status management
- Due dates

### Admin Features

- Admin authentication
- View registered users
- View tasks assigned to a specific user
- Update task status
- Create and assign tasks to any user

### Password Recovery

- Forgot password request
- Password reset email
- Secure password reset token
- Reset password through email link
- Password reset tokens expire after a limited time

## Project Structure

```text
task-manager/
│
├── config/
│   └── database.go
│
├── middleware/
│   ├── auth.go
│   └── admin.go
│
├── models/
│   ├── user.go
│   └── taskinput.go
│
├── utils/
│   ├── email.go
│   └── passwordreset.go
│
├── .env
├── .gitignore
├── go.mod
├── go.sum
├── main.go
└── README.md
