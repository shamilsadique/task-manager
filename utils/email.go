package utils

import (
	"fmt"
	"os"
	"github.com/go-mail/mail/v2"
)

func SendPasswordResetEmail(toEmail string, resetLink string) error {

	host := os.Getenv("SMTP_HOST")
	port := 587
	email := os.Getenv("SMTP_EMAIL")
	password := os.Getenv("SMTP_PASSWORD")

	message := mail.NewMessage()

	message.SetHeader("From", email)
	message.SetHeader("To", toEmail)
	message.SetHeader("Subject", "Password Reset - Task Manager")

	message.SetBody("text/html", fmt.Sprintf(`
		<h2>Password Reset</h2>

		<p>You requested to reset your Task Manager password.</p>

		<p>Click the button below to create a new password:</p>

		<p>
			<a href="%s"
			   style="
			   display:inline-block;
			   padding:10px 20px;
			   background:#2563eb;
			   color:white;
			   text-decoration:none;
			   border-radius:5px;">
				Reset Password
			</a>
		</p>

		<p>This link will expire in 15 minutes.</p>

		<p>If you did not request a password reset, you can safely ignore this email.</p>
	`, resetLink))

	dialer := mail.NewDialer(
		host,
		port,
		email,
		password,
	)

	return dialer.DialAndSend(message)
}