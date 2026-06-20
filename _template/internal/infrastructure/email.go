package infrastructure

import "github.com/resendlabs/resend-go"

type EmailService interface {
	Send(to, subject, html string) error
}

type ResendEmailService struct {
	client      *resend.Client
	fromAddress string
}

func NewResendEmailService(apiKey, fromAddress string) *ResendEmailService {
	return &ResendEmailService{
		client:      resend.NewClient(apiKey),
		fromAddress: fromAddress,
	}
}

func (s *ResendEmailService) Send(to, subject, html string) error {
	_, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.fromAddress,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	})
	return err
}
