package settings

import (
	"encoding/json"
	"fmt"
	"time"

	"goph/internal/infrastructure"
)

const emailTokenTTL = 1 * time.Hour

type emailConfirmData struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
}

func CreateEmailConfirmToken(secretKey, userID, email string) (string, error) {
	data := emailConfirmData{UserID: userID, Email: email}
	payload, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("marshal email confirm data: %w", err)
	}
	return infrastructure.CreateToken(string(payload), secretKey, emailTokenTTL)
}

func VerifyEmailConfirmToken(secretKey, token string) (*emailConfirmData, error) {
	payload, err := infrastructure.VerifyToken(token, secretKey)
	if err != nil {
		return nil, err
	}
	var data emailConfirmData
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		return nil, fmt.Errorf("unmarshal email confirm data: %w", err)
	}
	return &data, nil
}

type EmailService interface {
	Send(to, subject, html string) error
}

func SendConfirmationEmail(svc EmailService, to, confirmURL string) error {
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
