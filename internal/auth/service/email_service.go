package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/smtp"
	"passiontree/internal/config"
	"time"
)

type EmailService interface {
	SendVerificationEmail(to, token string) error
}

type emailServiceImpl struct {
	config *config.Config
}

func NewEmailService(cfg *config.Config) EmailService {
	return &emailServiceImpl{
		config: cfg,
	}
}

// SendVerificationEmail sends an email verification link
func (s *emailServiceImpl) SendVerificationEmail(to, token string) error {
	if s.config.SMTPHost == "" || s.config.SMTPFromEmail == "" {
		return fmt.Errorf("SMTP configuration is not set")
	}

	verificationURL := fmt.Sprintf("%s/auth/verify-email?token=%s", s.config.AppURL, token)

	subject := "Email Verification - Passion Tree"
	body := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 10px; text-align: center; }
        .content { padding: 20px; background-color: #f4f4f4; }
        .button { display: inline-block; padding: 10px 20px; background-color: #4CAF50; color: white; text-decoration: none; border-radius: 5px; }
        .footer { margin-top: 20px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Passion Tree</h1>
        </div>
        <div class="content">
            <h2>ยืนยันอีเมลของคุณ</h2>
            <p>ขอบคุณที่ลงทะเบียนกับ Passion Tree!</p>
            <p>กรุณาคลิกปุ่มด้านล่างเพื่อยืนยันอีเมลของคุณ:</p>
            <p style="text-align: center;">
                <a href="%s" class="button">ยืนยันอีเมล</a>
            </p>
            <p>หรือคัดลอกลิงก์นี้ไปวางในเบราว์เซอร์:</p>
            <p style="word-break: break-all;">%s</p>
            <p><strong>ลิงก์นี้จะหมดอายุใน 24 ชั่วโมง</strong></p>
        </div>
        <div class="footer">
            <p>หากคุณไม่ได้สร้างบัญชีนี้ กรุณาเพิกเฉยอีเมลนี้</p>
        </div>
    </div>
</body>
</html>
`, verificationURL, verificationURL)

	message := fmt.Sprintf("From: %s\r\n"+
		"To: %s\r\n"+
		"Subject: %s\r\n"+
		"MIME-Version: 1.0\r\n"+
		"Content-Type: text/html; charset=\"UTF-8\"\r\n"+
		"\r\n"+
		"%s\r\n", s.config.SMTPFromEmail, to, subject, body)

	auth := smtp.PlainAuth("", s.config.SMTPUsername, s.config.SMTPPassword, s.config.SMTPHost)
	addr := fmt.Sprintf("%s:%s", s.config.SMTPHost, s.config.SMTPPort)

	err := smtp.SendMail(addr, auth, s.config.SMTPFromEmail, []string{to}, []byte(message))
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	return nil
}

// GenerateVerificationToken generates a random token for email verification
func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// GetVerificationTokenExpiry returns expiry time (24 hours from now)
func GetVerificationTokenExpiry() time.Time {
	return time.Now().Add(24 * time.Hour)
}
