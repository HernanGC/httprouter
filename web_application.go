package httprouter

import (
	"net/http"
)

type WebApplication interface {
	Post(path string, handler http.HandlerFunc, mws ...Middleware)
	Get(path string, handler http.HandlerFunc, mws ...Middleware)
	Put(path string, handler http.HandlerFunc, mws ...Middleware)
	Patch(path string, handler http.HandlerFunc, mws ...Middleware)
	Delete(path string, handler http.HandlerFunc, mws ...Middleware)
}

type Middleware func(handlerFunc http.HandlerFunc) http.HandlerFunc
