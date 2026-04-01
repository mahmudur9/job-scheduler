package mocks

import "sync"

type MockJobQueue struct {
	mu sync.Mutex

	PublishFn func(body []byte, routingKey string) error

	PublishCalled int
	Messages      [][]byte
}

func (m *MockJobQueue) Publish(body []byte, routingKey string) error {
	m.mu.Lock()
	m.PublishCalled++
	m.Messages = append(m.Messages, body)
	m.mu.Unlock()

	return m.PublishFn(body, routingKey)
}
