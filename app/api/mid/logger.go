package mid

import (
	"context"
	"fmt"
	"time"

	"github.com/ardanlabs/service/foundation/logger"
	"github.com/ardanlabs/service/foundation/web"
)

// Logger writes information about the request to the log.
func Logger(ctx context.Context, log *logger.Logger, path string, rawQuery string, method string, remoteAddr string, handler Handler) error {
	v := web.GetValues(ctx)

	if rawQuery != "" {
		path = fmt.Sprintf("%s?%s", path, rawQuery)
	}

	log.Info(ctx, "requets started", "method", method, "path", path, "remoteAddr", remoteAddr)

	err := handler(ctx)

	log.Info(ctx, "requets completed", "method", method, "path", path, "remoteAddr", remoteAddr, "statuscode", v.StatusCode, "since", time.Since(v.Now).String())

	return err
}
