package utils

import (
	"strings"
	"testing"
)

func TestRenderEmailTemplate(t *testing.T) {
	rendered, err := RenderEmailTemplate("base_email.html", map[string]string{
		"Email":  "user@example.com",
		"Verify": "https://example.com/verify",
	})
	if err != nil {
		t.Fatalf("RenderEmailTemplate returned an error: %v", err)
	}

	if !strings.Contains(rendered, "user@example.com") {
		t.Fatalf("rendered body did not include the recipient email")
	}

	if !strings.Contains(rendered, "https://example.com/verify") {
		t.Fatalf("rendered body did not include the verification link")
	}
}
