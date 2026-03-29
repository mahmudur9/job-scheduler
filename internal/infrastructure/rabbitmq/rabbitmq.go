package rabbitmq

import (
	"fmt"
	"job-scheduler/internal/usecase"
	"log"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

const (
	reconnectDelay = 5 * time.Second
	maxRetries     = 7200
)

type RabbitMQ struct {
	url     string
	conn    *amqp091.Connection
	ch      *amqp091.Channel
	mu      sync.RWMutex
	closed  bool
	closeCh chan struct{}
}

// New creates a new RabbitMQ instance and establishes the initial connection.
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

// connect (re)establishes the AMQP connection and channel, then declares the
// queue. It is safe to call multiple times.
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
	if _, err = ch.QueueDeclare("jobs", true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return err
	}

	m.mu.Lock()
	m.conn = conn
	m.ch = ch
	m.mu.Unlock()

	// Watch for connection-level errors and trigger reconnect.
	go m.watchConnection(conn)
	return nil
}

// watchConnection blocks until the connection closes, then reconnects unless
// the RabbitMQ instance itself has been shut down.
func (m *RabbitMQ) watchConnection(conn *amqp091.Connection) {
	connErr := make(chan *amqp091.Error, 1)
	conn.NotifyClose(connErr)

	select {
	case err, ok := <-connErr:
		if !ok || err == nil {
			// Clean close — nothing to do.
			return
		}
		log.Printf("[rabbitmq] connection lost: %v — reconnecting…", err)
	case <-m.closeCh:
		return
	}

	m.reconnectLoop()
}

// reconnectLoop retries connect() with a fixed back-off until it succeeds or
// the RabbitMQ instance is closed.
func (m *RabbitMQ) reconnectLoop() {
	for attempt := 1; attempt <= maxRetries; attempt++ {
		select {
		case <-m.closeCh:
			return
		case <-time.After(reconnectDelay):
		}

		log.Printf("[rabbitmq] reconnect attempt %d/%d", attempt, maxRetries)
		if err := m.connect(); err != nil {
			log.Printf("[rabbitmq] reconnect attempt %d failed: %v", attempt, err)
			continue
		}
		log.Printf("[rabbitmq] reconnected successfully on attempt %d", attempt)
		return
	}
	log.Printf("[rabbitmq] giving up after %d reconnect attempts", maxRetries)
}

// channel returns the current channel under a read-lock.
func (m *RabbitMQ) channel() *amqp091.Channel {
	m.mu.RLock() // No one can modify it while reading. Multiple readers can read concurrently
	defer m.mu.RUnlock()
	return m.ch
}

// Consume starts consuming from queueName and feeds messages into the returned
// channel. When the underlying AMQP delivery channel closes (e.g. on
// disconnect), it waits for reconnection and re-subscribes automatically.
func (m *RabbitMQ) Consume(queueName string) (<-chan usecase.Message, error) {
	// Validate the initial subscription works before returning.
	if _, err := m.channel().Consume(queueName, "", false, false, false, false, nil); err != nil {
		return nil, err
	}

	out := make(chan usecase.Message)
	go m.consumeLoop(queueName, out)
	return out, nil
}

// consumeLoop re-subscribes whenever the delivery channel is closed and
// forwards messages to out until RabbitMQ.Close() is called.
func (m *RabbitMQ) consumeLoop(queueName string, out chan<- usecase.Message) {
	defer close(out)

	for {
		amqpMsgs, err := m.channel().Consume(queueName, "", false, false, false, false, nil)
		if err != nil {
			log.Printf("[rabbitmq] consume error: %v — waiting for reconnect", err)
			if !m.waitForReconnect() {
				return // MQ was closed
			}
			continue
		}

		// Forward messages until the delivery channel closes.
		for msg := range amqpMsgs {
			select {
			case out <- &AMQPMessage{msg}:
			case <-m.closeCh:
				return
			}
		}

		// amqpMsgs was closed — check if we should stop or reconnect.
		select {
		case <-m.closeCh:
			return
		default:
			log.Printf("[rabbitmq] delivery channel closed — waiting for reconnect")
			if !m.waitForReconnect() {
				return
			}
		}
	}
}

// waitForReconnect polls until the channel is healthy again or RabbitMQ is closed.
// Returns false when RabbitMQ.Close() has been called.
func (m *RabbitMQ) waitForReconnect() bool {
	for {
		select {
		case <-m.closeCh:
			return false
		case <-time.After(reconnectDelay):
			m.mu.RLock()
			ch := m.ch
			m.mu.RUnlock()
			if ch != nil && !ch.IsClosed() {
				return true
			}
		}
	}
}

// Publish sends body to the "jobs" queue. It retries once after a short delay
// if the channel is currently unavailable due to an ongoing reconnect.
func (m *RabbitMQ) Publish(body []byte) error {
	publish := func() error {
		m.mu.RLock()
		ch := m.ch
		m.mu.RUnlock()
		if ch == nil || ch.IsClosed() {
			return fmt.Errorf("channel not available")
		}
		return ch.Publish("", "jobs", false, false,
			amqp091.Publishing{
				DeliveryMode: amqp091.Persistent,
				Body:         body,
			})
	}

	if err := publish(); err != nil {
		log.Printf("[rabbitmq] publish failed: %v — retrying after delay", err)
		select {
		case <-m.closeCh:
			return err
		case <-time.After(reconnectDelay):
		}
		return publish()
	}
	return nil
}

// Close gracefully shuts down the RabbitMQ instance.
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

// AMQPMessage wraps an amqp091.Delivery to satisfy the usecase.Message interface.
type AMQPMessage struct {
	amqp091.Delivery
}

func (m *AMQPMessage) Body() []byte            { return m.Delivery.Body }
func (m *AMQPMessage) Ack() error              { return m.Delivery.Ack(false) }
func (m *AMQPMessage) Nack(requeue bool) error { return m.Delivery.Nack(false, requeue) }
