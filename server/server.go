package server

import "github.com/sirupsen/logrus"

type Server struct {
	Log *logrus.Logger
}

func InitServer(server *Server) *Server {
	return &Server{
		Log: server.Log,
	}
}
