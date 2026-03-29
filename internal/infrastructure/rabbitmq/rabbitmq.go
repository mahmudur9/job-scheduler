package rabbitmq

import (
	"errors"
	"fmt"
	"job-scheduler/internal/usecase"
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const (
	reconnectDelay = 5 * time.Second
)

type RabbitMQ struct {
	url     string
	conn    *amqp091.Connection
	ch      *amqp091.Channel
	mu      sync.RWMutex
	closed  bool
	closeCh chan struct{}
}

func New(url string) (*RabbitMQ, error) {
	mq := &RabbitMQ{
		url:     url,
		closeCh: make(chan struct{}),
	}
	if err := mq.connect(); err != nil {
		return nil, err
	}
	return mq, nil
}

func (m *RabbitMQ) connect() error {
	conn, err := amqp091.Dial(m.url)
	if err != nil {
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return err
	}

	if err := ch.Qos(1, 0, false); err != nil { // Each worker will do one job at a time.
		ch.Close()
		conn.Close()
		return err
	}

	_, err = ch.QueueDeclare("jobs", true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return err
	}

	// swap safely and close old resources
	m.mu.Lock()
	oldConn := m.conn
	oldCh := m.ch

	m.conn = conn
	m.ch = ch
	m.mu.Unlock()

	if oldCh != nil {
		oldCh.Close()
	}
	if oldConn != nil {
		oldConn.Close()
	}

	go m.watchConnection(conn)
	go m.watchChannel(ch)

	return nil
}

func (m *RabbitMQ) watchConnection(conn *amqp091.Connection) {
	errCh := make(chan *amqp091.Error, 1)
	conn.NotifyClose(errCh)

	select {
	case err := <-errCh:
		if err != nil {
			log.Printf("[rabbitmq] connection lost: %v", err)
			m.reconnectLoop()
		}
	case <-m.closeCh:
	}
}

func (m *RabbitMQ) watchChannel(ch *amqp091.Channel) {
	errCh := make(chan *amqp091.Error, 1)
	ch.NotifyClose(errCh)

	select {
	case err := <-errCh:
		if err != nil {
			log.Printf("[rabbitmq] channel closed: %v", err)
			m.reconnectLoop()
		}
	case <-m.closeCh:
	}
}

func (m *RabbitMQ) reconnectLoop() {
	for {
		select {
		case <-m.closeCh:
			return
		case <-time.After(reconnectDelay):
		}

		log.Println("[rabbitmq] reconnecting...")

		if err := m.connect(); err != nil {
			log.Printf("[rabbitmq] reconnect failed: %v", err)
			continue
		}

		log.Println("[rabbitmq] reconnected successfully")
		return
	}
}

// safer access
func (m *RabbitMQ) getChannel() (*amqp091.Channel, error) {
	m.mu.RLock() // No one will be able to modify the channel while accessing it
	defer m.mu.RUnlock()

	if m.ch == nil || m.ch.IsClosed() {
		return nil, errors.New("channel not available")
	}
	return m.ch, nil
}

func (m *RabbitMQ) Consume(queue string) (<-chan usecase.Message, error) {
	out := make(chan usecase.Message)
	go m.consumeLoop(queue, out)
	return out, nil
}

func (m *RabbitMQ) consumeLoop(queue string, out chan<- usecase.Message) {
	defer close(out)

	for {
		ch, err := m.getChannel()
		if err != nil {
			if !m.waitForReconnect() {
				return
			}
			continue
		}

		msgs, err := ch.Consume(queue, "", false, false, false, false, nil)
		if err != nil {
			log.Printf("[rabbitmq] consume error: %v", err)
			if !m.waitForReconnect() {
				return
			}
			continue
		}

		for msg := range msgs {
			select {
			case out <- &AMQPMessage{msg}:
			case <-m.closeCh:
				return
			}
		}
	}
}

func (m *RabbitMQ) waitForReconnect() bool {
	for {
		select {
		case <-m.closeCh:
			return false
		case <-time.After(reconnectDelay):
			m.mu.RLock()
			ok := m.ch != nil && !m.ch.IsClosed()
			m.mu.RUnlock()
			if ok {
				return true
			}
		}
	}
}

func (m *RabbitMQ) Publish(body []byte) error {
	for i := 0; i < 2; i++ {
		ch, err := m.getChannel()
		if err != nil {
			log.Printf("[rabbitmq] publish: channel unavailable")
		} else {
			err = ch.Publish(
				"",
				"jobs",
				false,
				false,
				amqp091.Publishing{
					DeliveryMode: amqp091.Persistent,
					Body:         body,
				},
			)
			if err == nil {
				return nil
			}
			log.Printf("[rabbitmq] publish failed: %v", err)
		}

		select {
		case <-m.closeCh:
			return fmt.Errorf("mq closed")
		case <-time.After(reconnectDelay):
		}
	}
	return fmt.Errorf("publish failed after retry")
}

func (m *RabbitMQ) Close() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.closed {
		return
	}
	m.closed = true
	close(m.closeCh)

	if m.ch != nil {
		m.ch.Close()
	}
	if m.conn != nil {
		m.conn.Close()
	}
}

type AMQPMessage struct {
	amqp091.Delivery
}

func (m *AMQPMessage) Body() []byte            { return m.Delivery.Body }
func (m *AMQPMessage) Ack() error              { return m.Delivery.Ack(false) }
func (m *AMQPMessage) Nack(requeue bool) error { return m.Delivery.Nack(false, requeue) }
