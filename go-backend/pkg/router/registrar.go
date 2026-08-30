package router

import "net/http"

// Registrar collects or mounts HTTP routes (Gin or go-zero).
type Registrar interface {
	Add(method, path string, h http.HandlerFunc)
}
