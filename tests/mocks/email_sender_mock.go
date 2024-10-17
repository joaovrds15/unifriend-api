package mocks

type MockSesSender struct {
	SendVerificationEmailFunc func(recipient, subject, body string) error
}

func (m *MockSesSender) SendVerificationEmail(recipient, subject, body string) error {
	return m.SendVerificationEmailFunc(recipient, subject, body)
}
