package server

import "github.com/sirupsen/logrus"

type Server struct {
	Clients *AtomicClientsMap
	Log     *logrus.Logger
}

func InitServer(
	clients *AtomicClientsMap,
	log *logrus.Logger,
) *Server {
	return &Server{
		Clients: clients,
		Log:     log,
	}
}
