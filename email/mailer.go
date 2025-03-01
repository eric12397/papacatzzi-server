package email

import (
	"bytes"
	"text/template"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	*gomail.Dialer
}

type EmailContent struct {
	Subject   string
	Recipient string
	Body      map[string]string
}

func NewMailer() Mailer {
	dialer := gomail.NewDialer("smtp.gmail.com", 587, "eric12397@gmail.com", "vkbe rtau koqr anwc")

	return Mailer{
		Dialer: dialer,
	}
}

func (m Mailer) Send(templatePath string, content EmailContent) error {
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return err
	}

	var body bytes.Buffer
	if err := tmpl.Execute(&body, content.Body); err != nil {
		return err
	}

	message := gomail.NewMessage()

	message.SetHeader("From", "eric12397@gmail.com")
	message.SetHeader("To", content.Recipient)
	message.SetHeader("Subject", content.Subject)
	message.SetBody("text/html", body.String())

	return m.DialAndSend(message)
}
