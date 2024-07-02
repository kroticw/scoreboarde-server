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

var timeDur = 20000

func (p *Play) Playing(ctx context.Context, stop context.CancelFunc) {
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
					},
				}
			} else {
				le := p.history.GetLast()
				lastEvent = *le
			}
			var event server.Message
			switch choise {
			case 1:
				event = server.Message{
					Time: lastEvent.Time + (time.Now().Unix() - lastEvent.Time),
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
						TimeInPeriod: time.Now().Unix() - lastEvent.Period.TimeStart,
					},
				}
				break
			case 2:
				event = server.Message{
					Time: lastEvent.Time + (time.Now().Unix() - lastEvent.Time),
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
						TimeInPeriod: time.Now().Unix() - lastEvent.Period.TimeStart,
					},
				}
				break
			case 3:
				time.Sleep(10000 * time.Millisecond)
				event = server.Message{
					Time: lastEvent.Time,
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
						TimeInPeriod: lastEvent.Period.TimeInPeriod,
					},
				}
				break
			default:
				<-ctx.Done()
			}
			if event.Period.TimeInPeriod >= p.PeriodDuration {
				event.Period.Count++
				event.Period.TimeStart = time.Now().Unix()
				event.Period.TimeInPeriod = 0
			}
			p.history.Push(event)
			err = p.server.Mailing(ctx, &event)
			if err != nil {
				p.Log.WithFields(logrus.Fields{
					"error": err,
				}).Errorln("error mailing")
				return
			}
		}

	}
}
