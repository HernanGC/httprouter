package httprouter

import (
	"net/http"
)

type methodHandlerMap map[string]http.HandlerFunc

type Application struct {
	mux                  *http.ServeMux
	mws                  []Middleware
	pathMethodHandlerMap map[string]methodHandlerMap
}

func NewApplication(mux *http.ServeMux) WebApplication {
	return &Application{
		mux:                  mux,
		pathMethodHandlerMap: make(map[string]methodHandlerMap),
	}
}

func (a *Application) WithGlobalMiddlewares(mws ...Middleware) WebApplication {
	a.mws = append(a.mws, mws...)
	return a
}

func (a *Application) handle(method string, path string, handler http.HandlerFunc, mws ...Middleware) {
	if a.pathMethodHandlerMap[path] == nil {
		a.pathMethodHandlerMap[path] = make(methodHandlerMap)

		a.mux.HandleFunc(path, func(w http.ResponseWriter, r *http.Request) {
			pMH := a.pathMethodHandlerMap[path]
			if h, ok := pMH[r.Method]; ok {
				h(w, r)
				return
			}
			w.Header().Set("Allow", a.allowedMethods(path))
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		})
	}

	a.pathMethodHandlerMap[path][method] = a.middleware(handler, mws...)
}

func (a *Application) allowedMethods(path string) string {
	methods := ""
	for m := range a.pathMethodHandlerMap[path] {
		if methods != "" {
			methods += ", "
		}
		methods += m
	}
	return methods
}

func (a *Application) middleware(handler http.HandlerFunc, mws ...Middleware) http.HandlerFunc {
	// First, apply route-specific middleware
	result := a.applyMiddleware(handler, mws...)
	// Then apply global middleware
	return a.applyMiddleware(result, a.mws...)
}

func (a *Application) applyMiddleware(handler http.HandlerFunc, mws ...Middleware) http.HandlerFunc {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}

func (a *Application) Post(path string, handler http.HandlerFunc, mws ...Middleware) {
	a.handle(http.MethodPost, path, handler, mws...)
}

func (a *Application) Get(path string, handler http.HandlerFunc, mws ...Middleware) {
	a.handle(http.MethodGet, path, handler, mws...)
}

func (a *Application) Put(path string, handler http.HandlerFunc, mws ...Middleware) {
	a.handle(http.MethodPut, path, handler, mws...)
}

func (a *Application) Patch(path string, handler http.HandlerFunc, mws ...Middleware) {
	a.handle(http.MethodPatch, path, handler, mws...)
}

func (a *Application) Delete(path string, handler http.HandlerFunc, mws ...Middleware) {
	a.handle(http.MethodDelete, path, handler, mws...)
}
