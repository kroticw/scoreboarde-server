package server

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"scoreboarde-server/Play"
)

func (s *Server) HandleTunnelClient(
	ctx context.Context,
	c *Client,
	history *AtomicMessageHistory,
	clients *AtomicClientsMap,
) {
	s.Log.WithFields(logrus.Fields{
		"con": c.ClientIp + ":" + c.ClientPort,
	}).Println("Connecting to client")

	_, exist := clients.GetClient(c.ClientIp)
	s.Log.Println("exist: ", exist)
	if !exist {
		clients.AddClient(c)
		for i := 0; i < history.Len(); i++ {
			err := s.SendToClient(ctx, c, history.Get(i))
			if err != nil {
				s.Log.WithFields(logrus.Fields{
					"con": c.ClientIp + ":" + c.ClientPort,
					"err": err,
				}).Error("Send to client error")
			}
		}
		s.Log.WithFields(logrus.Fields{
			"ip": c.ClientIp,
		}).Println("Create new client")
	}
	defer c.Close(clients)
	for {
		com1 := &Play.Command{
			Name:  "one",
			Score: 10,
		}
		com2 := &Play.Command{
			Name:  "two",
			Score: 5,
		}
		per := &Play.Period{
			Count:        1,
			TimeInPeriod: 0,
		}
		mes := &Message{
			Time:       0,
			CommandOne: *com1,
			CommandTwo: *com2,
			Period:     *per,
		}
		err := s.SendToClient(ctx, c, mes)
		if err != nil {
			s.Log.WithFields(logrus.Fields{
				"ip":    c.ClientIp,
				"error": err,
			}).Errorln("Send message to client")
			return
		}
	}
}

func (s *Server) SendToClient(_ context.Context, c *Client, mes *Message) error {
	s.Log.WithFields(logrus.Fields{
		"con": c.ClientIp + ":" + c.ClientPort,
		"mes": mes,
	}).Println("Send to client")

	err := json.NewEncoder(c.Conn).Encode(mes)
	if err != nil {
		s.Log.WithFields(logrus.Fields{
			"con":   c.ClientIp + ":" + c.ClientPort,
			"mes":   mes,
			"error": err,
		}).Errorln("error send")
		return err
	}

	return nil
}
