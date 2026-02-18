package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"net/smtp"
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
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 1. ลองส่งด้วย MailerSend ก่อน (ถ้าตั้งค่า API Key ไว้)
	if s.config.MailerSendAPIKey != "" {
		message := s.mailersendClient.Email.NewMessage()
		message.SetFrom(mailersend.From{Name: fromName, Email: s.config.SMTPFromEmail})
		message.SetRecipients([]mailersend.Recipient{{Name: "User", Email: to}})
		message.SetSubject(subject)
		message.SetHTML(html)
		message.SetText(text)

		_, err := s.mailersendClient.Email.Send(ctx, message)
		if err == nil {
			s.logger.Info("email sent via MailerSend", "to", to, "subject", subject)
			return nil
		}

		// ถ้า MailerSend พัง หรือยังไม่ Approve (มักจะได้ Error กลับมา) ให้ Log ไว้แล้วไปต่อที่ Gmail
		s.logger.Warn("MailerSend failed, trying fallback to Gmail", "error", err)
	}

	// 2. Fallback: ส่งด้วย Gmail SMTP (ถ้า MailerSend ใช้ไม่ได้)
	return s.sendViaGmail(to, subject, html)
}

func (s *emailServiceImpl) sendViaGmail(to, subject, htmlBody string) error {
	from := s.config.GmailEmail
	password := s.config.GmailAppPassword

	// 1. สร้างหัวจดหมาย (Headers) - ต้องใช้ \r\n เท่านั้น
	header := fmt.Sprintf("From: %s\r\n", from)
	header += fmt.Sprintf("To: %s\r\n", to)
	header += fmt.Sprintf("Subject: %s\r\n", subject)
	header += "MIME-Version: 1.0\r\n"
	header += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	header += "\r\n" // บรรทัดว่างแบ่ง Header และ Body (ห้ามขาด!)

	// 2. รวมร่าง Message
	message := []byte(header + htmlBody)

	// 3. Authentication & Send
	auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
	err := smtp.SendMail("smtp.gmail.com:587", auth, from, []string{to}, message)
	
	if err != nil {
		s.logger.Error("Gmail ultimate failure", "error", err)
		return err
	}

	s.logger.Info("OTP sent via Gmail successfully", "to", to)
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