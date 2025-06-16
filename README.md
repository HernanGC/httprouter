# HttpRouter

A lightweight, flexible HTTP routing abstraction for Go that works with the standard library's `http.ServeMux`.

[![Go Reference](https://pkg.go.dev/badge/github.com/HernanGC/httprouter.svg)](https://pkg.go.dev/github.com/HernanGC/httprouter)
[![Go Report Card](https://goreportcard.com/badge/github.com/HernanGC/httprouter)](https://goreportcard.com/report/github.com/HernanGC/httprouter)

## Overview
Initial implementation of HttpRouter

A lightweight HTTP routing abstraction for Go that builds on top of the standard
library's http.ServeMux. This implementation provides:

- Core routing functionality with support for HTTP methods (GET, POST, PUT, PATCH, DELETE)
- Middleware support with simple composition pattern
- Automatic handling of 405 Method Not Allowed responses
- Clean interface design through WebApplication abstraction
- No external dependencies

The implementation consists of two main components:
- WebApplication interface defining the routing contract
- Application struct implementing the routing logic with method-based handlers

This provides a solid foundation for building HTTP applications with a focus on
simplicity and compatibility with Go's standard library.

## Features

- HTTP method-based routing (GET, POST, PUT, PATCH, DELETE)
- Middleware support with simple composition
- Automatic 405 Method Not Allowed responses with proper Allow headers
- Compatible with standard Go HTTP handlers
- No external dependencies
- Built on top of Go's standard `http.ServeMux`

## Installation

```bash
go get github.com/HernanGC/httprouter
```

## Quick Start

```go
package main

import (
	"fmt"
	"net/http"

	"github.com/HernanGC/httprouter"
)

func main() {
	// Create a new ServeMux
	mux := http.NewServeMux()

	// Create a new application with the mux
	app := httprouter.NewApplication(mux)

	// Define routes with HTTP methods
	app.Get("/", indexHandler)
	app.Get("/users", listUsersHandler)
	app.Post("/users", createUserHandler)
	app.Get("/users/profile", getUserProfileHandler, authMiddleware)

	// Start the server
	fmt.Println("Server started at http://localhost:8080")
	http.ListenAndServe(":8080", mux)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Welcome to the home page!")
}

func listUsersHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "List of users")
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "User created")
}

func getUserProfileHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "User profile")
}

// Example middleware
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Check for auth header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		// Continue to the next handler if authorized
		next(w, r)
	}
}
```

## Core Concepts

### WebApplication Interface

The `WebApplication` interface defines the contract for HTTP method-based routing:

```go
type WebApplication interface {
	Post(path string, handler http.HandlerFunc, mws ...Middleware)
	Get(path string, handler http.HandlerFunc, mws ...Middleware)
	Put(path string, handler http.HandlerFunc, mws ...Middleware)
	Patch(path string, handler http.HandlerFunc, mws ...Middleware)
	Delete(path string, handler http.HandlerFunc, mws ...Middleware)
}
```

### Middleware

Middleware functions allow you to execute code before and after your handlers:

```go
type Middleware func(handlerFunc http.HandlerFunc) http.HandlerFunc
```

Middlewares are applied in the order they are provided, with the last middleware in the list being executed first.

## Examples

### Using Multiple Middlewares

```go
app.Get("/protected", protectedHandler, loggingMiddleware, authMiddleware, rateLimitMiddleware)
```

### Creating Common Middleware

```go
func loggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next(w, r)
		fmt.Printf("%s %s %s %v\n", r.Method, r.URL.Path, r.RemoteAddr, time.Since(start))
	}
}

func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next(w, r)
	}
}
```

## Security Considerations

This router provides a convenient abstraction for HTTP routing but does not include built-in security features. When using this package, please consider implementing:

- Input validation and sanitization
- Authentication and authorization
- CSRF protection
- Rate limiting
- Input/output encoding

These security concerns should be addressed in your application code or middleware.

## Compatibility

HttpRouter is compatible with Go 1.24 and later versions.

## License

This project is licensed under the Apache-2.0 license - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
