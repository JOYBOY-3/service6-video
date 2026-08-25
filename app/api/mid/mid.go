// Package mid provides app level middleware support
package mid

import "context"

type Handler func(context.Context) error
