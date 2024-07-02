package cmd

import (
	"context"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"net"
	"os"
	"os/signal"
	"scoreboarde-server/Play"
	"scoreboarde-server/server"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "WebSocket Транслятор игры",
	Long: `Запускает веб-сервер для прослушки UDP порта и обработки вебхуков TCP для клиентов.
			Отдаёт информацию об игре в текущий момент времени`,
	Run: executeServeCommand,
}

func init() {
	rootCmd.AddCommand(serveCmd)
}

func executeServeCommand(_ *cobra.Command, _ []string) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	clients := server.NewAtomicClientsMap()

	s := server.InitServer(
		clients,
		logger,
	)

	historyPlay := server.InitAtomicMessageHistory()
	p := Play.InitPlay(
		cfg.CommandOne,
		cfg.CommandTwo,
		cfg.PeriodDuration,
		historyPlay,
		s,
		logger,
	)

	serverAddress, err := net.ResolveUDPAddr("udp4", "192.168.88.255:8889")
	if err != nil {
		logger.WithFields(logrus.Fields{
			"err": err,
		}).Println("resolve udp address error")
		return
	}
	connection, err := net.ListenUDP("udp", serverAddress)
	if err != nil {
		logger.WithFields(logrus.Fields{
			"error": err,
		}).Println("start listen error")
		return
	}
	defer connection.Close()
	go func() {
		for {
			select {
			case <-ctx.Done():
				stop()
				return
			default:
				err := s.UdpConnectionListener(ctx, connection, historyPlay, clients)
				if err != nil {
					logger.WithFields(logrus.Fields{
						"error": err,
					}).Errorln("UdpConnectionListener error")
					return
				}
			}
		}
	}()

	p.Playing(ctx, stop)
}
