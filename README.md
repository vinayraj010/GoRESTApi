Here is the updated `README.md` without icons, without the License section, and with detailed explanations of the code, methods, and the exact flow of how the application works, including the code for all the files.

```markdown
# GoRESTApi - Student Management API

A simple and robust RESTful API built with Go (Golang) to manage student records. It uses SQLite for data storage, `cleanenv` for configuration management, and includes graceful shutdown support.

## Project Structure

```text
GoRESTApi/
├── cmd/
│   └── student-api/
│       └── main.go          # Application entry point
├── config/
│   └── local.yaml           # Local configuration file
├── internal/
│   ├── config/              # Configuration loader
│   ├── http/
│   │   └── handlers/
│   │       └── student/     # HTTP handlers for student routes
│   ├── storage/
│   │   ├── sqlite/          # SQLite database implementation
│   │   └── storage.go       # Storage interface
│   ├── types/               # Data models (Structs)
│   └── utils/
│       └── response/        # Standardized HTTP JSON responses
├── storage/                 # SQLite database files (auto-generated)
├── go.mod
├── go.sum
└── README.md
```

## Prerequisites

- Go (Version 1.22 or higher recommended, since we use the new `net/http` routing features like `GET /api/students/{id}`).
- Git

## Setup & Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/vinayraj010/GoRESTApi.git
   cd GoRESTApi
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Create the storage directory:
   The application needs a folder to store the SQLite database file.
   ```bash
   mkdir storage
   ```

4. Ensure you have a `config/local.yaml` file in the `config` directory. It should look like this:
   ```yaml
   env: "dev"
   storage: "storage/storage.db"
   http_server:
     address: "localhost:8082"
   ```

5. Run the application:
   ```bash
   go run cmd/student-api/main.go -config config/local.yaml
   ```

   You should see logs indicating the storage is initialized and the server has started:
   ```text
   2026/11/22 06:30:00 INFO storage initialized env=dev version=1.0.0
   2026/11/22 06:30:00 INFO server started address=localhost:8082
   ```

## Code Flow and File Explanation

This section explains how the application works internally, file by file, and how the flow of a request is handled.

### 1. `config/local.yaml` and `internal/config/config.go`
The application starts by reading the configuration file. The `local.yaml` file holds the environment variables, database path, and the HTTP server address. The `config.go` file uses the `cleanenv` package to parse this YAML file into a Go struct.

```go
// internal/config/config.go
package config

import (
	"flag"
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env        string     `yaml:"env" env:"ENV" env-required:"true"`
	Storage    string     `yaml:"storage" env:"STORAGE" env-required:"true"`
	HTTPServer HTTPServer `yaml:"http_server"`
}

type HTTPServer struct {
	Addr string `yaml:"address" env:"HTTP_ADDR" env-required:"true"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		configFlag := flag.String("config", "", "path to config file")
		flag.Parse()
		configPath = *configFlag
		if configPath == "" {
			log.Fatal("config file path is not specified")
		}
	}
	if _, err := os.Stat(configPath); err != nil {
		log.Fatalf("config file error: %s", err)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("failed to read config file: %s", err)
	}
	return &cfg
}
```

### 2. `internal/types/types.go`
This file defines the core data structure for the application. The `Student` struct represents a student record. The `validate` tags are used by the `go-playground/validator` package to ensure that required fields are present when creating or updating a student.

```go
// internal/types/types.go
package types

type Student struct {
	Id    int    `json:"id"`
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required"`
	Age   int    `json:"age" validate:"required"`
}
```

### 3. `internal/storage/storage.go`
This file defines an interface named `Storage`. This is a crucial design pattern in Go. By defining an interface, we decouple the HTTP handlers from the actual database implementation. If you ever want to switch from SQLite to PostgreSQL, you only need to create a new implementation of this interface without touching the handlers.

```go
// internal/storage/storage.go
package storage

import "github.com/vinayraj010/GoRESTApi/internal/types"

type Storage interface {
	CreateStudent(name string, email string, age int) (int64, error)
	GetStudentById(id int64) (types.Student, error)
	GetStudents() ([]types.Student, error)
	UpdateStudent(id int64, name string, email string, age int) error
	DeleteStudent(id int64) error
}
```

### 4. `internal/storage/sqlite/sqlite.go`
This is the concrete implementation of the `Storage` interface. It uses the `database/sql` package and the `mattn/go-sqlite3` driver to interact with the SQLite database.

- `New(cfg *config.Config)`: Opens the database connection and creates the `students` table if it doesn't exist.
- `CreateStudent`: Uses an `INSERT` statement to add a new student.
- `GetStudentById`: Uses a `SELECT` statement with a `LIMIT 1` to fetch a single student.
- `GetStudents`: Uses a `SELECT` statement to fetch all students and iterates through the rows.
- `UpdateStudent`: Uses an `UPDATE` statement to modify an existing student's details.
- `DeleteStudent`: Uses a `DELETE` statement to remove a student.

```go
// internal/storage/sqlite/sqlite.go
package sqlite

import (
	"database/sql"

	"github.com/vinayraj010/GoRESTApi/internal/config"
	"github.com/vinayraj010/GoRESTApi/internal/types"
	_ "github.com/mattn/go-sqlite3"
)

type Sqlite struct {
	Db *sql.DB
}

func New(cfg *config.Config) (*Sqlite, error) {
	db, err := sql.Open("sqlite3", cfg.Storage)
	if err != nil {
		return nil, err
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS students (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		email TEXT,
		age INTEGER
	)`)
	if err != nil {
		return nil, err
	}

	return &Sqlite{Db: db}, nil
}

func (s *Sqlite) CreateStudent(name string, email string, age int) (int64, error) {
	stmt, err := s.Db.Prepare("INSERT INTO students (name, email, age) VALUES (?, ?, ?)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	result, err := stmt.Exec(name, email, age)
	if err != nil {
		return 0, err
	}

	return result.LastInsertId()
}

func (s *Sqlite) GetStudentById(id int64) (types.Student, error) {
	stmt, err := s.Db.Prepare("SELECT * FROM students WHERE id = ? LIMIT 1")
	if err != nil {
		return types.Student{}, err
	}
	defer stmt.Close()

	var student types.Student
	err = stmt.QueryRow(id).Scan(&student.Id, &student.Name, &student.Email, &student.Age)
	if err != nil {
		return types.Student{}, err
	}

	return student, nil
}

func (s *Sqlite) GetStudents() ([]types.Student, error) {
	stmt, err := s.Db.Prepare("SELECT id, name, email, age FROM students")
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []types.Student
	for rows.Next() {
		var student types.Student
		err = rows.Scan(&student.Id, &student.Name, &student.Email, &student.Age)
		if err != nil {
			return nil, err
		}
		students = append(students, student)
	}

	return students, nil
}

func (s *Sqlite) UpdateStudent(id int64, name string, email string, age int) error {
	stmt, err := s.Db.Prepare("UPDATE students SET name = ?, email = ?, age = ? WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(name, email, age, id)
	return err
}

func (s *Sqlite) DeleteStudent(id int64) error {
	stmt, err := s.Db.Prepare("DELETE FROM students WHERE id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(id)
	return err
}
```

### 5. `internal/utils/response/response.go`
This utility file standardizes the JSON responses sent back to the client. It uses a `Response` struct to keep the format consistent (always having a `status` and an optional `error`). The `WriteJson` function handles setting the headers and encoding the data.

```go
// internal/utils/response/response.go
package response

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
)

type Response struct {
	Status string      `json:"status"`
	Error  interface{} `json:"error,omitempty"`
}

const (
	StatusOK    = "OK"
	StatusError = "Error"
)

func WriteJson(w http.ResponseWriter, status int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(data)
}

func GeneralError(err error) Response {
	return Response{
		Status: StatusError,
		Error:  err.Error(),
	}
}

func ValidationError(errs validator.ValidationErrors) Response {
	var errMsgs []string
	for _, err := range errs {
		switch err.ActualTag() {
		case "required":
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is required field", err.Field()))
		default:
			errMsgs = append(errMsgs, fmt.Sprintf("field %s is invalid", err.Field()))
		}
	}
	return Response{
		Status: StatusError,
		Error:  errMsgs,
	}
}
```

### 6. `internal/http/handlers/student/student.go`
This file contains the logic for handling HTTP requests. Each function acts as a handler for a specific route.

**Request Flow Example (Create Student):**
1. The `New` handler receives the HTTP request.
2. It decodes the JSON body into a `types.Student` struct.
3. It validates the struct using the `validator` package.
4. If validation passes, it calls `storage.CreateStudent()`.
5. It logs the success and returns the new ID to the user.

Similar flows occur for Get, Update, and Delete, utilizing `r.PathValue("id")` to extract the ID from the URL path.

```go
// internal/http/handlers/student/student.go
package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/vinayraj010/GoRESTApi/internal/storage"
	"github.com/vinayraj010/GoRESTApi/internal/types"
	"github.com/vinayraj010/GoRESTApi/internal/utils/response"
)

func New(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("creating a student")

		var student types.Student
		err := json.NewDecoder(r.Body).Decode(&student)
		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if err := validator.New().Struct(student); err != nil {
			validateErrs := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}

		lastId, err := storage.CreateStudent(student.Name, student.Email, student.Age)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		slog.Info("user created successfully", slog.String("userId", fmt.Sprint(lastId)))
		response.WriteJson(w, http.StatusCreated, map[string]int64{"id": lastId})
	}
}

func GetById(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("getting a student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("invalid id")))
			return
		}

		student, err := storage.GetStudentById(intId)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, student)
	}
}

func GetList(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("getting all students")

		students, err := storage.GetStudents()
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, students)
	}
}

func Update(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("invalid id")))
			return
		}

		var student types.Student
		err = json.NewDecoder(r.Body).Decode(&student)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if err := validator.New().Struct(student); err != nil {
			validateErrs := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}

		err = storage.UpdateStudent(intId, student.Name, student.Email, student.Age)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, map[string]string{"message": "student updated successfully"})
	}
}

func Delete(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("invalid id")))
			return
		}

		err = storage.DeleteStudent(intId)
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, map[string]string{"message": "student deleted successfully"})
	}
}
```

### 7. `cmd/student-api/main.go`
This is the entry point of the application. It ties everything together.

**Application Startup Flow:**
1. It loads the configuration using `config.MustLoad()`.
2. It initializes the SQLite database by calling `sqlite.New(cfg)`.
3. It sets up the HTTP router (`http.NewServeMux`) and registers the endpoints, passing the `storage` instance to the handlers.
4. It starts the HTTP server in a separate goroutine so it doesn't block the main thread.
5. It listens for OS interrupts (`Ctrl+C`) using the `os/signal` package.
6. Upon receiving a signal, it initiates a graceful shutdown with a 5-second timeout using `context.WithTimeout`.

```go
// cmd/student-api/main.go
package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/vinayraj010/GoRESTApi/internal/config"
	"github.com/vinayraj010/GoRESTApi/internal/http/handlers/student"
	"github.com/vinayraj010/GoRESTApi/internal/storage/sqlite"
)

func main() {
	cfg := config.MustLoad()

	storage, err := sqlite.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	slog.Info("storage initialized", slog.String("env", cfg.Env), slog.String("version", "1.0.0"))

	router := http.NewServeMux()

	router.HandleFunc("POST /api/students", student.New(storage))
	router.HandleFunc("GET /api/students/{id}", student.GetById(storage))
	router.HandleFunc("GET /api/students", student.GetList(storage))
	router.HandleFunc("PUT /api/students/{id}", student.Update(storage))
	router.HandleFunc("DELETE /api/students/{id}", student.Delete(storage))

	server := http.Server{
		Addr:    cfg.HTTPServer.Addr,
		Handler: router,
	}

	slog.Info("server started", slog.String("address", cfg.HTTPServer.Addr))

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		err := server.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			log.Fatalf("failed to start server: %v", err)
		}
	}()

	<-done

	slog.Info("shutting down the server")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("failed to shutdown server", slog.String("error", err.Error()))
	}

	slog.Info("server shutdown successfully")
}
```

## API Endpoints

| Method | Endpoint | Description |
| :--- | :--- | :--- |
| POST | `/api/students` | Create a new student |
| GET | `/api/students` | Get a list of all students |
| GET | `/api/students/{id}` | Get details of a specific student by ID |
| PUT | `/api/students/{id}` | Update an existing student's details |
| DELETE | `/api/students/{id}` | Delete a student by ID |

### Request/Response Examples

1. Create a Student (POST `/api/students`)
Request Body:
```json
{
  "name": "John Doe",
  "email": "john@example.com",
  "age": 20
}
```
Response (201 Created):
```json
{
  "id": 1
}
```

2. Get All Students (GET `/api/students`)
Response (200 OK):
```json
[
  {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com",
    "age": 20
  }
]
```

## Stopping the Server
To stop the server gracefully, press `Ctrl + C` in the terminal where it is running. The server will catch the interrupt signal, close the database connection, and shut down cleanly.
```
