package Play

import (
	"context"
	"scoreboarde-server/server"
)

type Client interface {
	SendToClient(ctx context.Context, mes *server.Message) error
}

type Server interface {
	Mailing(ctx context.Context, mes *server.Message) error
}
