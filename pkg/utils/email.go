package utils

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"time"
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
	apiKey := os.Getenv("BREVO_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("BREVO API KEY isn't set yet.")
	}

	payload, _ := json.Marshal(map[string]any{
		"sender": map[string]string{
			"email": os.Getenv("BREVO_SENDER_EMAIL"),
			"name":  os.Getenv("BREVO_SENDER_NAME"),
		},
		"to":          []map[string]string{{"email": toEmail}},
		"subject":     subject,
		"htmlContent": body,
	})

	req, _ := http.NewRequest("POST", "https://api.brevo.com/v3/smtp/email", bytes.NewBuffer(payload))
	req.Header.Set("api-key", apiKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode >= 300 {
		b, _ := io.ReadAll(res.Body)
		return fmt.Errorf("Brevo: status %d - %s", res.StatusCode, b)
	}

	return nil
}
