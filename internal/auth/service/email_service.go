package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"passiontree/internal/pkg/apperror"

	"github.com/mailersend/mailersend-go"
)

func (s *emailServiceImpl) SendVerificationEmail(to, token string) error {
	subject := "รหัสยืนยันตัวตน - Passion-Tree"
	text := fmt.Sprintf("ยินดีต้อนรับสู่ Passion-Tree!\n\nรหัสยืนยันตัวตนของคุณคือ: %s\n\nรหัสนี้จะหมดอายุใน 15 นาที", token)

	// Execute template with token data
	var htmlBuf bytes.Buffer
	if err := s.templates.verification.Execute(&htmlBuf, map[string]string{"Token": token}); err != nil {
		s.logger.Error("failed to execute verification template", "error", err)
		return apperror.NewInternal("failed to generate email content")
	}

	return s.sendEmail(to, subject, htmlBuf.String(), text, "Passiontree Team")
}

func (s *emailServiceImpl) SendPasswordResetEmail(to, token string) error {
	subject := "รีเซ็ตรหัสผ่าน - Passiontree"
	text := fmt.Sprintf("รหัสรีเซ็ตรหัสผ่าน: %s\n\nรหัสนี้จะหมดอายุใน 15 นาที", token)

	// Execute template with token data
	var htmlBuf bytes.Buffer
	if err := s.templates.passwordReset.Execute(&htmlBuf, map[string]string{"Token": token}); err != nil {
		s.logger.Error("failed to execute password reset template", "error", err)
		return apperror.NewInternal("failed to generate email content")
	}

	return s.sendEmail(to, subject, htmlBuf.String(), text, "Passiontree Security")
}

func (s *emailServiceImpl) SendSecurityAlertEmail(to, userID string) error {
	subject := "แจ้งเตือนความปลอดภัย - Passiontree"
	now := time.Now().Format("02/01/2006 15:04:05")
	text := fmt.Sprintf("ตรวจพบกิจกรรมที่น่าสงสัยในบัญชีของคุณ\n\nUser ID: %s\nเวลา: %s\n\nเราได้ล็อกเอาท์ทุกอุปกรณ์เพื่อความปลอดภัย\nกรุณาเปลี่ยนรหัสผ่านทันที", userID, now)

	// Execute template with security alert data
	var htmlBuf bytes.Buffer
	data := map[string]string{
		"UserID":    userID,
		"Timestamp": now,
	}
	if err := s.templates.securityAlert.Execute(&htmlBuf, data); err != nil {
		s.logger.Error("failed to execute security alert template", "error", err)
		return apperror.NewInternal("failed to generate email content")
	}

	return s.sendEmail(to, subject, htmlBuf.String(), text, "Passiontree Security Team")
}

func (s *emailServiceImpl) sendEmail(to, subject, html, text, fromName string) error {
	if s.config.MailerSendAPIKey == "" || s.config.SMTPFromEmail == "" {
		return apperror.NewInternal("email service not configured")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message := s.mailersendClient.Email.NewMessage()
	message.SetFrom(mailersend.From{Name: fromName, Email: s.config.SMTPFromEmail})
	message.SetRecipients([]mailersend.Recipient{{Name: "User", Email: to}})
	message.SetSubject(subject)
	message.SetHTML(html)
	message.SetText(text)

	if _, err := s.mailersendClient.Email.Send(ctx, message); err != nil {
		s.logger.Error("send email failed", "to", to, "error", err)
		return apperror.NewInternal("failed to send email: %w", err)
	}

	s.logger.Info("email sent", "to", to, "subject", subject)
	return nil
}

func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	code := (int(bytes[0])<<24|int(bytes[1])<<16|int(bytes[2])<<8|int(bytes[3]))%900000 + 100000
	return fmt.Sprintf("%06d", code), nil
}

func GetVerificationTokenExpiry() time.Time {
	return time.Now().Add(15 * time.Minute)
}
