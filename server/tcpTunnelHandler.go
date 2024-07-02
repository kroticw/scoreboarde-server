package server

import (
	"context"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
)

func (s *Server) HandleTunnelClient(
	ctx context.Context,
	c *Client,
	history *AtomicMessageHistory,
) {
	s.Log.WithFields(logrus.Fields{
		"con": c.ClientIp + ":" + c.ClientPort,
	}).Println("Handle to client")

	_, exist := s.Clients.GetClient(c.ClientIp)
	s.Log.Println("exist: ", exist)
	if !exist {
		s.Clients.AddClient(c)
		for i := 0; i < history.Len(); i++ {
			err := c.SendToClient(ctx, history.Get(i))
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
	defer c.Close(s.Clients)
	for {

	}
}

func (c *Client) SendToClient(_ context.Context, mes *Message) error {
	c.log.WithFields(logrus.Fields{
		"con": c.ClientIp + ":" + c.ClientPort,
		"mes": mes,
	}).Println("Send to client")

	b, err := json.Marshal(mes)
	if err != nil {
		return err
	}

	bs := []byte(strconv.Itoa(len(b)))
	//err = binary.Write(bs, binary.LittleEndian, uint32(len(b)))
	//if err != nil {
	//	fmt.Println("binary.Write failed:", err)
	//}
	_, err = c.Conn.Write(bs)
	if err != nil {
		return err
	}

	c.log.Infoln(bs, "bytes sent; ", uint32(len(b)))

	time.Sleep(100 * time.Millisecond)
	err = json.NewEncoder(c.Conn).Encode(mes)
	if err != nil {
		c.log.WithFields(logrus.Fields{
			"con":   c.ClientIp + ":" + c.ClientPort,
			"mes":   mes,
			"error": err,
		}).Errorln("error send")
		return err
	}

	return nil
}

func (s *Server) Mailing(_ context.Context, mes *Message) error {
	mapForSending := s.Clients.clientsMap
	for _, client := range mapForSending {
		err := client.SendToClient(context.Background(), mes)
		if err != nil {
			return err
		}
	}
	return nil
}
