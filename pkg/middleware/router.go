package middleware

import (
	"net/http"
	"strings"

	"github.com/julienschmidt/httprouter"
)

// CustomRouter extends httprouter.Router to handle API routes separately from data routes
type CustomRouter struct {
	*httprouter.Router
	apiRouter  *httprouter.Router
	dataRouter *httprouter.Router
}

// NewCustomRouter creates a new CustomRouter
func NewCustomRouter() *CustomRouter {
	return &CustomRouter{
		Router:     httprouter.New(),
		apiRouter:  httprouter.New(),
		dataRouter: httprouter.New(),
	}
}

// ServeHTTP implements the http.Handler interface
func (r *CustomRouter) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// Check if the request is for the API
	if strings.HasPrefix(req.URL.Path, "/__/") {
		r.apiRouter.ServeHTTP(w, req)
		return
	}

	// Otherwise, use the data router
	r.dataRouter.ServeHTTP(w, req)
}

// GET registers a GET handler for the API router if path starts with /__/, otherwise for the data router
func (r *CustomRouter) GET(path string, handle httprouter.Handle) {
	if strings.HasPrefix(path, "/__/") {
		r.apiRouter.GET(path, handle)
	} else {
		r.dataRouter.GET(path, handle)
	}
}

// POST registers a POST handler for the API router if path starts with /__/, otherwise for the data router
func (r *CustomRouter) POST(path string, handle httprouter.Handle) {
	if strings.HasPrefix(path, "/__/") {
		r.apiRouter.POST(path, handle)
	} else {
		r.dataRouter.POST(path, handle)
	}
}

// PATCH registers a PATCH handler for the API router if path starts with /__/, otherwise for the data router
func (r *CustomRouter) PATCH(path string, handle httprouter.Handle) {
	if strings.HasPrefix(path, "/__/") {
		r.apiRouter.PATCH(path, handle)
	} else {
		r.dataRouter.PATCH(path, handle)
	}
}

// PUT registers a PUT handler for the API router if path starts with /__/, otherwise for the data router
func (r *CustomRouter) PUT(path string, handle httprouter.Handle) {
	if strings.HasPrefix(path, "/__/") {
		r.apiRouter.PUT(path, handle)
	} else {
		r.dataRouter.PUT(path, handle)
	}
}

// DELETE registers a DELETE handler for the API router if path starts with /__/, otherwise for the data router
func (r *CustomRouter) DELETE(path string, handle httprouter.Handle) {
	if strings.HasPrefix(path, "/__/") {
		r.apiRouter.DELETE(path, handle)
	} else {
		r.dataRouter.DELETE(path, handle)
	}
}

// OPTIONS registers an OPTIONS handler for the API router if path starts with /__/, otherwise for the data router
func (r *CustomRouter) OPTIONS(path string, handle httprouter.Handle) {
	if strings.HasPrefix(path, "/__/") {
		r.apiRouter.OPTIONS(path, handle)
	} else {
		r.dataRouter.OPTIONS(path, handle)
	}
}
