package Play

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"scoreboarde-server/server"
	"time"
)

type Play struct {
	TimeStart      int64
	Time           int64
	CommandOne     server.Command
	CommandTwo     server.Command
	PeriodDuration int64
	history        *server.AtomicMessageHistory
	server         *server.Server
	Log            *logrus.Logger
}

func InitPlay(
	commandOneName string,
	commandTwoName string,
	periodDuration int64,
	history *server.AtomicMessageHistory,
	s *server.Server,
	log *logrus.Logger,
) *Play {
	commandOne := server.Command{
		Name:  commandOneName,
		Score: 0,
	}
	commandTwo := server.Command{
		Name:  commandTwoName,
		Score: 0,
	}
	return &Play{
		TimeStart:      time.Now().UnixMilli(),
		Time:           0,
		CommandOne:     commandOne,
		CommandTwo:     commandTwo,
		PeriodDuration: periodDuration,
		history:        history,
		server:         s,
		Log:            log,
	}
}

var messChan = make(chan server.Message, 1)

func (p *Play) Queue(ctx context.Context) {
	for {
		if len(messChan) == 0 {
			lastEvent := p.history.GetLast()
			var mes server.Message
			mes.Time = lastEvent.Time
			mes.CommandOne.Name = p.CommandOne.Name
			mes.CommandOne.Score = lastEvent.CommandOne.Score
			mes.CommandTwo.Name = p.CommandTwo.Name
			mes.CommandTwo.Score = lastEvent.CommandTwo.Score
			mes.Period.TimeInPeriod = lastEvent.Period.TimeInPeriod
			mes.Period.Count = lastEvent.Period.Count
			mes.Period.Pause = lastEvent.Period.Pause
			if !lastEvent.Period.Pause {
				mes.Time = lastEvent.Time + 60
				mes.Period.TimeInPeriod = lastEvent.Period.TimeInPeriod + 60
			}

			if mes.Period.TimeInPeriod >= p.PeriodDuration {
				mes.Period.Count++
				mes.Period.TimeStart = time.Now().Unix()
				mes.Period.TimeInPeriod = 0
			}
			err := p.server.Mailing(ctx, &mes)
			if err != nil {
				logrus.WithError(err).Errorln("error queuing message")

			}
			p.history.Push(mes)
		} else {
			message := <-messChan
			err := p.server.Mailing(ctx, &message)
			if err != nil {
				logrus.WithError(err).Errorln("error queuing message from chan")

			}
		}
		time.Sleep(time.Second * 1)
	}
}

func (p *Play) Playing(ctx context.Context, stop context.CancelFunc) {

	fmt.Println("Start? [yn]: ")
	var ch string
	_, _ = fmt.Scanln(&ch)
	switch ch {
	case "y":
		var lastEvent server.Message
		if p.history.Len() == 0 {
			lastEvent = server.Message{
				Time: 0,
				CommandOne: server.Command{
					Name:  p.CommandOne.Name,
					Score: p.CommandOne.Score,
				},
				CommandTwo: server.Command{
					Name:  p.CommandTwo.Name,
					Score: p.CommandTwo.Score,
				},
				Period: server.Period{
					TimeStart:    time.Now().Unix(),
					Count:        1,
					TimeInPeriod: 0,
					Pause:        false,
				},
			}
			p.history.Push(lastEvent)
		}
		time.Sleep(1 * time.Second)
		go p.Queue(ctx)
		for {
			select {
			case <-ctx.Done():
			default:
				fmt.Printf(
					"Выберите ситуацию:\n"+
						"1. Команда %s забивает гол\n"+
						"2. Команда %s забивает гол\n"+
						"3. Time!\n",
					p.CommandOne.Name,
					p.CommandTwo.Name,
				)
				var choise int
				_, err := fmt.Scanf("%d\n", &choise)
				if err != nil {
					p.Log.WithFields(logrus.Fields{
						"error": err,
					}).Errorln("Ошибка ввода игровой ситуации")
					continue
				}

				le := p.history.GetLast()
				lastEvent = *le
				var event server.Message
				switch choise {
				case 1:
					event = server.Message{
						Time: lastEvent.Time + 60,
						CommandOne: server.Command{
							Name:  p.CommandOne.Name,
							Score: lastEvent.CommandOne.Score + 1,
						},
						CommandTwo: server.Command{
							Name:  p.CommandTwo.Name,
							Score: lastEvent.CommandTwo.Score,
						},
						Period: server.Period{
							Count:        lastEvent.Period.Count,
							TimeInPeriod: lastEvent.Period.TimeInPeriod + 1,
							Pause:        false,
						},
					}
					break
				case 2:
					event = server.Message{
						Time: lastEvent.Time + 1,
						CommandOne: server.Command{
							Name:  p.CommandOne.Name,
							Score: lastEvent.CommandOne.Score,
						},
						CommandTwo: server.Command{
							Name:  p.CommandTwo.Name,
							Score: lastEvent.CommandTwo.Score + 1,
						},
						Period: server.Period{
							Count:        lastEvent.Period.Count,
							TimeInPeriod: lastEvent.Period.TimeInPeriod + 60,
							Pause:        false,
						},
					}
					break
				case 3:
					event = server.Message{
						Time: lastEvent.Time,
						CommandOne: server.Command{
							Name:  p.CommandOne.Name,
							Score: lastEvent.CommandOne.Score,
						},
						CommandTwo: server.Command{
							Name:  p.CommandTwo.Name,
							Score: lastEvent.CommandTwo.Score,
						},
						Period: server.Period{
							Count:        lastEvent.Period.Count,
							TimeInPeriod: lastEvent.Period.TimeInPeriod,
							Pause:        !lastEvent.Period.Pause,
						},
					}
					break
				case 12:
					<-ctx.Done()
				default:

				}
				if event.Period.TimeInPeriod >= p.PeriodDuration {
					event.Period.Count++
					event.Period.TimeStart = time.Now().Unix()
					event.Period.TimeInPeriod = 0
				}
				p.history.Push(event)
				messChan <- event
			}
		}
	default:
		<-ctx.Done()
	}

}
