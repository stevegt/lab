package v2

import (
	"context"
)

// Module interface with unified Accept and HandleMessage
type Module interface {
	Accept(ctx context.Context, parms ...interface{}) (Message, error)
	HandleMessage(ctx context.Context, parms ...interface{}) ([]byte, error)
}
