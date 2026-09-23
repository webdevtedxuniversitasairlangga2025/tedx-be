package handler

import "testing"

func TestProofHostAllowed(t *testing.T) {
	allowed := []string{"ik.imagekit.io", "upload.imagekit.io", "cdn.imagekit.io"}
	blocked := []string{
		"internetpositif.id",
		"evil.com",
		"imagekit.io.evil.com",
		"localhost",
		"169.254.169.254",
		"",
	}
	for _, h := range allowed {
		if !proofHostAllowed(h) {
			t.Errorf("proofHostAllowed(%q) = false, want true", h)
		}
	}
	for _, h := range blocked {
		if proofHostAllowed(h) {
			t.Errorf("proofHostAllowed(%q) = true, want false", h)
		}
	}
}
