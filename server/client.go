package server

import (
	"github.com/sirupsen/logrus"
	"net"
	"sync"
)

type Client struct {
	ClientIp   string `json:"client_ip"`
	ClientPort string `json:"client_port"`
	Conn       *net.TCPConn
	log        *logrus.Logger
}

type AtomicClientsMap struct {
	mu         sync.Mutex
	clientsMap map[string]*Client
}

func NewAtomicClientsMap() *AtomicClientsMap {
	return &AtomicClientsMap{
		clientsMap: make(map[string]*Client),
	}
}

func (a *AtomicClientsMap) AddClient(client *Client) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.clientsMap[client.ClientIp] = client
}

func (a *AtomicClientsMap) RemoveClient(clientIp string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	delete(a.clientsMap, clientIp)

}

func (a *AtomicClientsMap) GetClient(clientIp string) (*Client, bool) {
	a.mu.Lock()
	defer a.mu.Unlock()
	client, ok := a.clientsMap[clientIp]
	return client, ok
}

func InitClient(client *Client) (*Client, error) {
	rtcpAddr, err := net.ResolveTCPAddr("tcp", client.ClientIp+":"+client.ClientPort)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"client_ip": client.ClientIp,
			"err":       err,
		}).Errorln("Error resolving tcp address")
		return &Client{}, err
	}
	tcpConn, err := net.DialTCP("tcp", nil, rtcpAddr)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"client_ip": client.ClientIp,
			"err":       err,
		}).Errorln("Error connecting to client")
		return &Client{}, err
	}

	return &Client{
		ClientIp:   client.ClientIp,
		ClientPort: client.ClientPort,
		Conn:       tcpConn,
		log:        client.log,
	}, nil
}

func (c *Client) Close(clients *AtomicClientsMap) {
	clients.RemoveClient(c.ClientIp)
	c.log.Infof("Removing client %s", c.ClientIp)
	c.Conn.Close()
}
