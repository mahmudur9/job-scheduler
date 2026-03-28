package rabbitmq

import "github.com/rabbitmq/amqp091-go"

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

func (m *MQ) Publish(body []byte) error {
	return m.ch.Publish("", "jobs", false, false,
		amqp091.Publishing{Body: body})
}
