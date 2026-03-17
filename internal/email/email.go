package email

import (
	"fmt"
	"html"

	"github.com/resend/resend-go/v3"
)

type Client struct {
	resend    *resend.Client
	fromEmail string
	baseURL   string
}

func New(apiKey, fromEmail, baseURL string) *Client {
	return &Client{
		resend:    resend.NewClient(apiKey),
		fromEmail: fromEmail,
		baseURL:   baseURL,
	}
}

func (c *Client) VerificationURL(code string) string {
	return c.baseURL + "/verify?code=" + code
}

func (c *Client) SendVerification(toEmail, username, verificationURL string) error {
	safeUsername := html.EscapeString(username)
	body := fmt.Sprintf(`<p>Hey %s,</p>
						 <p>Thanks for signing up! Please verify your email address:</p>
						 <p><a href="%s">Verify Email</a></p>
						 <p>This link expires in 24 hours. 
						 	If you didn't create this account, you can safely ignore this email.</p>`,
		safeUsername, verificationURL)

	_, err := c.resend.Emails.Send(&resend.SendEmailRequest{
		From:    c.fromEmail,
		To:      []string{toEmail},
		Subject: "Verify your email address",
		Html:    body,
	})
	if err != nil {
		return fmt.Errorf("failed to send verification email: %w", err)
	}
	return nil
}
