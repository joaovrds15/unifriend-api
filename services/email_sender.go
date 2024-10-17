package services

type SesSender interface {
	SendVerificationEmail(recipient, subject, body string) error
}
