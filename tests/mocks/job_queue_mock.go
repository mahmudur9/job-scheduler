package mocks

import "sync"

type MockJobQueue struct {
	mu sync.Mutex

	PublishFn func(body []byte) error

	PublishCalled int
	Messages      [][]byte
}

func (m *MockJobQueue) Publish(body []byte) error {
	m.mu.Lock()
	m.PublishCalled++
	m.Messages = append(m.Messages, body)
	m.mu.Unlock()

	return m.PublishFn(body)
}
