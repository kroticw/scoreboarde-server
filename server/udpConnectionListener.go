package server

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"net"
	"strings"
	"time"
)

func (s *Server) UdpConnectionListener(
	ctx context.Context,
	connection *net.UDPConn,
	history *AtomicMessageHistory,
	clients *AtomicClientsMap,
) error {
	err := connection.SetReadDeadline(time.Now().Add(1 * time.Second))
	if err != nil {
		s.Log.WithFields(logrus.Fields{
			"error": err,
		}).Println("set read deadline error")
		return err
	}

	var newClient Client
	err = json.NewDecoder(connection).Decode(&newClient)
	if err != nil {
		if !strings.Contains(err.Error(), "i/o timeout") {
			s.Log.WithFields(logrus.Fields{
				"error": err,
			}).Errorln("Decode error")
			return err
		}
		return nil
	}

	s.Log.WithFields(logrus.Fields{
		"newClient": newClient,
	}).Println("Received message")

	c, err := InitClient(&Client{
		ClientIp:   newClient.ClientIp,
		ClientPort: newClient.ClientPort,
		log:        s.Log,
	})
	if err != nil {
		s.Log.WithFields(logrus.Fields{
			"error": err,
		}).Errorln("InitClient error")
	}
	go s.HandleTunnelClient(ctx, c, history, clients)
	return nil
}
