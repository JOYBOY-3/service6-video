// Package web contains a small web framework extension.

package web

import (
	"context"
	"errors"
	"net/http"
	"os"
	"syscall"
	"time"

	"github.com/google/uuid"
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

// SignalShutdown is used to gracefully shut down the app when integrity issue is identified
func (a *App) SignalShutdown() {
	a.shutdown <- syscall.SIGTERM
}

// HandleFunc sets a handler fuction for a given HTTP method and path pair
// to the application server mux
func (a *App) HandleFunc(pattern string, handler Handler, mw ...MidHandler) {
	handler = wrapMiddleware(mw, handler)
	handler = wrapMiddleware(a.mw, handler)

	h := func(w http.ResponseWriter, r *http.Request) {
		v := Values{
			TraceID: uuid.NewString(),
			Now:     time.Now().UTC(),
		}

		ctx := setValues(r.Context(), &v)

		if err := handler(ctx, w, r); err != nil {
			if validateError(err) {
				a.SignalShutdown()
				return
			}
		}

	}

	a.ServeMux.HandleFunc(pattern, h)
}

// HandleFuncNoMiddleware sets a handler fuction for a given HTTP method
// and path pair to the application server mux with no middleware
func (a *App) HandleFuncNoMiddleware(pattern string, handler Handler, mw ...MidHandler) {
	h := func(w http.ResponseWriter, r *http.Request) {
		v := Values{
			TraceID: uuid.NewString(),
			Now:     time.Now().UTC(),
		}

		ctx := setValues(r.Context(), &v)

		if err := handler(ctx, w, r); err != nil {
			if validateError(err) {
				a.SignalShutdown()
				return
			}
		}

	}

	a.ServeMux.HandleFunc(pattern, h)
}

// validateError validates the error for special conditions that do not warrant an actudal shutdown by the system.
func validateError(err error) bool {
	// Ignore syscall.EPIPE and syscall.ECONNRESET errors whcih occurs
	// when a write operation happens on the http.ResponseWriter that
	// has simultaneously been disconnected by the client (TCP
	// connections is broken). For instance, when large amounts
	// of data is being written a streamed to the client.
	// https:/blog.cloudflaare.com/the-complete--guide-to-golang-net-http-timouts
	// https://gosamples.dev/connection-reset-by-peer

	switch {
	case errors.Is(err, syscall.EPIPE):

		// Usually, you get the broken pipe error when you erite to the connection after the
		// TST (TCP RST Flag) is sent.
		// The broken pipe is a TCP/IP error occuring when you write to a stream where the
		// other end (the peer) has closed the underlyig connection. The first write to the
		// closd connection should be terminated immediately. The second write to the socket that
		//  has already reccieved the RST cuases the broken pip peer
		return false

	case errors.Is(err, syscall.ECONNRESET):
		// Ususually, you get connection reset by peer error when you read from the
		// connection after the RST (TCP RST Flag) is sent.
		// The conection reset by peer s a TCP/IP erro that occurs when the other en (peer)
		// has unexpectedly closed the connection. It happens when you send a packet from your
		// end, but the other end crashes and forcibly closes the connection under normal
		// circcmstances
		return false

	}

	return true
}
