package rabbitmq

import (
	"job-scheduler/internal/usecase"

	"github.com/rabbitmq/amqp091-go"
)

type MQ struct {
	conn *amqp091.Connection
	ch   *amqp091.Channel
}

func New(url string) (*MQ, error) {
	conn, _ := amqp091.Dial(url)
	ch, _ := conn.Channel()

	ch.QueueDeclare("jobs", true, false, false, false, nil)

	return &MQ{conn, ch}, nil
}

func (m *MQ) Channel() *amqp091.Channel {
	return m.ch
}
func (r *MQ) Consume(queueName string) (<-chan usecase.Message, error) {
	amqpMsgs, err := r.ch.Consume(queueName, "", false, false, false, false, nil)
	if err != nil {
		return nil, err
	}

	out := make(chan usecase.Message)
	go func() {
		defer close(out)
		for m := range amqpMsgs {
			out <- &AMQPMessage{m}
		}
	}()
	return out, nil
}

type AMQPMessage struct {
	amqp091.Delivery
}

func (m *AMQPMessage) Body() []byte            { return m.Delivery.Body }
func (m *AMQPMessage) Ack() error              { return m.Delivery.Ack(false) }
func (m *AMQPMessage) Nack(requeue bool) error { return m.Delivery.Nack(false, requeue) }
func (m *MQ) Publish(body []byte) error {
	return m.ch.Publish("", "jobs", false, false,
		amqp091.Publishing{Body: body})
}
