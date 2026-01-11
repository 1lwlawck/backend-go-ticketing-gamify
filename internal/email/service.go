package email

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

// Service handles email sending via Resend
type Service struct {
	apiKey      string
	fromEmail   string
	fromName    string
	frontendURL string
}

// NewService creates a new email service from environment variables
func NewService() *Service {
	return &Service{
		apiKey:      os.Getenv("RESEND_API_KEY"),
		fromEmail:   getEnv("RESEND_FROM_EMAIL", "noreply@resend.dev"),
		fromName:    getEnv("RESEND_FROM_NAME", "Ticketing Gamified"),
		frontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// IsConfigured checks if Resend is properly configured
func (s *Service) IsConfigured() bool {
	return s.apiKey != ""
}

// ResendRequest represents the Resend API request body
type ResendRequest struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	HTML    string   `json:"html"`
}

// SendVerificationEmail sends email verification link via Resend
func (s *Service) SendVerificationEmail(to, name, token, frontendURL string) error {
	if !s.IsConfigured() {
		// Log but don't fail - useful for development
		fmt.Printf("[EMAIL] Would send verification email to %s with token %s\n", to, token)
		fmt.Printf("[EMAIL] Verification URL: %s/verify-email?token=%s\n", frontendURL, token)
		return nil
	}

	verifyURL := fmt.Sprintf("%s/verify-email?token=%s", frontendURL, token)

	htmlBody := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
  <style>
    body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; line-height: 1.6; color: #333; margin: 0; padding: 0; background: #f4f4f5; }
    .container { max-width: 600px; margin: 0 auto; padding: 40px 20px; }
    .card { background: white; border-radius: 16px; padding: 40px; box-shadow: 0 4px 6px rgba(0,0,0,0.05); }
    .header { text-align: center; margin-bottom: 30px; }
    .header h1 { color: #4F46E5; margin: 0; font-size: 28px; }
    .content { margin-bottom: 30px; }
    .button { display: inline-block; background: linear-gradient(135deg, #4F46E5 0%%, #7C3AED 100%%); color: white !important; text-decoration: none; padding: 16px 40px; border-radius: 12px; font-weight: 600; font-size: 16px; }
    .button-container { text-align: center; margin: 30px 0; }
    .link { color: #4F46E5; word-break: break-all; font-size: 14px; }
    .footer { text-align: center; color: #71717a; font-size: 13px; margin-top: 30px; padding-top: 20px; border-top: 1px solid #e4e4e7; }
    .badge { display: inline-block; background: #fef3c7; color: #92400e; padding: 4px 12px; border-radius: 20px; font-size: 12px; margin-top: 10px; }
  </style>
</head>
<body>
  <div class="container">
    <div class="card">
      <div class="header">
        <h1>🎮 Ticketing Gamified</h1>
      </div>
      <div class="content">
        <h2 style="color: #18181b; margin-top: 0;">Hi %s! 👋</h2>
        <p>Welcome to <strong>Ticketing Gamified</strong>! You're just one step away from starting your journey.</p>
        <p>Please verify your email address to complete your registration and unlock all features.</p>
        <div class="button-container">
          <a href="%s" class="button">✨ Verify Email Address</a>
        </div>
        <p style="font-size: 14px; color: #71717a;">Or copy and paste this link in your browser:</p>
        <p class="link">%s</p>
        <span class="badge">⏰ This link expires in 24 hours</span>
      </div>
      <div class="footer">
        <p>If you didn't create an account, you can safely ignore this email.</p>
        <p>© 2026 Ticketing Gamified. All rights reserved.</p>
      </div>
    </div>
  </div>
</body>
</html>
`, name, verifyURL, verifyURL)

	reqBody := ResendRequest{
		From:    fmt.Sprintf("%s <%s>", s.fromName, s.fromEmail),
		To:      []string{to},
		Subject: "🔐 Verify Your Email - Ticketing Gamified",
		HTML:    htmlBody,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("resend API error (status %d): %s", resp.StatusCode, string(body))
	}

	fmt.Printf("[EMAIL] Verification email sent to %s via Resend\n", to)
	return nil
}
