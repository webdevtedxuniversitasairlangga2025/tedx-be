package utils

import (
	"bytes"
	"embed"
	"html/template"

	"github.com/webdevtedxuniversitasairlangga/config"
	"gopkg.in/gomail.v2"
)

//go:embed email-template/*.html
var emailTemplates embed.FS

func RenderEmailTemplate(templateName string, data map[string]string) (string, error) {
	tpl, err := template.ParseFS(emailTemplates, "email-template/"+templateName)
	if err != nil {
		return "", err
	}

	var body bytes.Buffer
	if err := tpl.Execute(&body, data); err != nil {
		return "", err
	}

	return body.String(), nil
}

func SendMail(toEmail string, subject string, body string) error {
	emailConfig, err := config.NewEmailConfig()
	if err != nil {
			return err
	}

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", emailConfig.AuthEmail)
	mailer.SetHeader("To", toEmail)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/html", body)

	dialer := gomail.NewDialer(
			emailConfig.Host,
			emailConfig.Port,
			emailConfig.AuthEmail,
			emailConfig.AuthPassword,
	)

	return dialer.DialAndSend(mailer)
}