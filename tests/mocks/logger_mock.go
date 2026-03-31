package mocks

type MockLogger struct {
	InfoFn  func(msg string)
	ErrorFn func(err error, msg string)
}

func (m *MockLogger) Info(msg string) {
	m.InfoFn(msg)
}

func (m *MockLogger) Error(err error, msg string) {
	m.ErrorFn(err, msg)
}
