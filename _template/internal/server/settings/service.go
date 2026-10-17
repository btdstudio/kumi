package settings

import (
	"goph/internal/infrastructure"
)

func SendConfirmationEmail(svc infrastructure.EmailService, to, confirmURL string) error {
	if svc == nil {
		return nil
	}
	return svc.Send(to, "Confirm your new email address",
		"<p>Hi,</p>"+
			"<p>Click the link below to confirm your new email address. This link expires in 1 hour.</p>"+
			"<p><a href=\""+confirmURL+"\">"+confirmURL+"</a></p>"+
			"<p>If you didn't request this, you can safely ignore this email.</p>",
	)
}
