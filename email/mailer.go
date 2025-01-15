package email

import (
	"fmt"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	*gomail.Dialer
}

func NewMailer() Mailer {
	dialer := gomail.NewDialer("smtp.gmail.com", 587, "eric12397@gmail.com", "vkbe rtau koqr anwc")

	return Mailer{
		Dialer: dialer,
	}
}

func (m Mailer) SendVerificationCode(recipient string, code string) error {
	// Create a new message
	message := gomail.NewMessage()

	// Set email headers
	message.SetHeader("From", "eric12397@gmail.com")
	message.SetHeader("To", recipient)
	message.SetHeader("Subject", "Your verification code")

	// Set email body
	text := fmt.Sprintf("Enter your verification code to finish signing up your new account: %s", code)
	message.SetBody("text/plain", text)

	// Send the email
	return m.DialAndSend(message)
}
