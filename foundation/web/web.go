// Package web contains a small web framework extension.

package web

import (
	"context"
	"fmt"
	"net/http"
	"os"
)

// A handler is a type that handles a http request within our own little mini framework
type Handler func(ctx context.Context, w http.ResponseWriter, r *http.Request) error

// App is the entrypoint into our application and what configures our context
// objetc for each of our http hanlders. Feel free to add any configuration
// data/logic on this App struct
type App struct {
	*http.ServeMux
	shutdown chan os.Signal
	mw       []MidHandler
}

// NewApp creates an App value that handle a set of routes for the application.
func NewApp(shutdown chan os.Signal, mw ...MidHandler) *App {
	return &App{
		ServeMux: http.NewServeMux(),
		shutdown: shutdown,
		mw:       mw,
	}
}

// Handle sets a handler fuction for a given HTTP method and path pair
// to the application server mux
func (a *App) HandleFunc(pattern string, handler Handler, mw ...MidHandler) {
	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {

		if err := handler(r.Context(), w, r); err != nil {
			// ERROR HANDLING HERE
			fmt.Println(err)
			return
		}

	}

	a.ServeMux.HandleFunc(pattern, h)
}
