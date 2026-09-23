package utils

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"

	"github.com/webdevtedxuniversitasairlangga/config"
	"gopkg.in/gomail.v2"
)

//go:embed email-template/*.html
var emailTemplates embed.FS

func RenderEmailTemplate[T any](templateName string, data T) (string, error) {
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
	mailer.SetHeader("From", emailConfig.SenderName)
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

func SendMailWithEmbeds(toEmail string, subject string, body string, embeds map[string][]byte) error {
	emailConfig, err := config.NewEmailConfig()
	if err != nil {
		return err
	}

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", emailConfig.SenderName)
	mailer.SetHeader("To", toEmail)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/html", body)

	for filename, data := range embeds {
		contentID := fmt.Sprintf("<%s>", filename)

		mailer.Attach(filename,
			gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(data)
				return err
			}),
			gomail.SetHeader(map[string][]string{
				"Content-ID":          {contentID},
				"Content-Type":        {"image/png; name=" + filename},
				"Content-Disposition": {fmt.Sprintf(`inline; filename="%s"`, filename)},
			}),
		)
	}

	dialer := gomail.NewDialer(
		emailConfig.Host,
		emailConfig.Port,
		emailConfig.AuthEmail,
		emailConfig.AuthPassword,
	)

	return dialer.DialAndSend(mailer)
}
