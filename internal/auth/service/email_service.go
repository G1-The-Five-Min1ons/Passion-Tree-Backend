package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"net/smtp"
	"strings"
	"time"

	"passiontree/internal/pkg/apperror"

	"github.com/mailersend/mailersend-go"
)

var smtpSendMail = smtp.SendMail

// ExportSetSMTPSendMail exposes smtpSendMail for tests
func ExportSetSMTPSendMail(mock func(addr string, a smtp.Auth, from string, to []string, msg []byte) error) func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
	original := smtpSendMail
	smtpSendMail = mock
	return original
}

// sanitizeHeaderValue removes newlines and carriage returns to prevent SMTP header injection
func sanitizeHeaderValue(value string) string {
	value = strings.ReplaceAll(value, "\r", "")
	value = strings.ReplaceAll(value, "\n", "")
	return value
}

func (s *emailServiceImpl) SendVerificationEmail(ctx context.Context, to, token string) error {
	subject := "รหัสยืนยันตัวตน - Passion-Tree"
	text := fmt.Sprintf("ยินดีต้อนรับสู่ Passion-Tree!\n\nรหัสยืนยันตัวตนของคุณคือ: %s\n\nรหัสนี้จะหมดอายุใน 15 นาที", token)

	// Execute template with token data
	var htmlBuf bytes.Buffer
	if err := s.templates.verification.Execute(&htmlBuf, map[string]string{"Token": token}); err != nil {
		s.logger.Error("failed to execute verification template", "error", err)
		return apperror.NewInternal("failed to generate email content")
	}

	return s.sendEmail(ctx, to, subject, htmlBuf.String(), text, "Passiontree Team")
}

func (s *emailServiceImpl) SendPasswordResetEmail(ctx context.Context, to, token string) error {
	subject := "รีเซ็ตรหัสผ่าน - Passiontree"
	text := fmt.Sprintf("รหัสรีเซ็ตรหัสผ่าน: %s\n\nรหัสนี้จะหมดอายุใน 15 นาที", token)

	// Execute template with token data
	var htmlBuf bytes.Buffer
	if err := s.templates.passwordReset.Execute(&htmlBuf, map[string]string{"Token": token}); err != nil {
		s.logger.Error("failed to execute password reset template", "error", err)
		return apperror.NewInternal("failed to generate email content")
	}

	return s.sendEmail(ctx, to, subject, htmlBuf.String(), text, "Passiontree Security")
}

func (s *emailServiceImpl) SendSecurityAlertEmail(ctx context.Context, to, userID string) error {
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

	return s.sendEmail(ctx, to, subject, htmlBuf.String(), text, "Passiontree Security Team")
}

func (s *emailServiceImpl) SendNotificationEmail(ctx context.Context, to, subject, headline, message string) error {
	text := fmt.Sprintf("%s\n\n%s", headline, message)
	html := fmt.Sprintf(`
		<div style="font-family: Arial, sans-serif; color:#1f2937; line-height:1.5;">
		  <h2 style="margin:0 0 8px 0;">%s</h2>
		  <p style="margin:0;">%s</p>
		</div>
	`, headline, message)

	return s.sendEmail(ctx, to, subject, html, text, "Passiontree Notifications")
}

func (s *emailServiceImpl) sendEmail(ctx context.Context, to, subject, html, text, fromName string) error {
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
	return s.sendViaGmail(ctx, to, subject, html, text)
}

func (s *emailServiceImpl) sendViaGmail(ctx context.Context, to, subject, htmlBody, textBody string) error {
	from := s.config.GmailEmail
	password := s.config.GmailAppPassword

	// Sanitize header values to prevent SMTP header injection
	to = sanitizeHeaderValue(to)
	subject = sanitizeHeaderValue(subject)

	// Generate boundary for multipart/alternative
	boundaryBuffer := make([]byte, 16)
	_, err := rand.Read(boundaryBuffer)
	if err != nil {
		return fmt.Errorf("failed to generate boundary: %w", err)
	}
	boundary := fmt.Sprintf("%x", boundaryBuffer)

	// 1. สร้างหัวจดหมาย (Headers) - ต้องใช้ \r\n เท่านั้น
	header := fmt.Sprintf("From: %s\r\n", from)
	header += fmt.Sprintf("To: %s\r\n", to)
	header += fmt.Sprintf("Subject: %s\r\n", subject)
	header += "MIME-Version: 1.0\r\n"
	header += fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary)
	header += "\r\n" // บรรทัดว่างแบ่ง Header และ Body (ห้ามขาด!)

	// 2. สร้าง Body ที่มีทั้ง Plain Text และ HTML Parts
	body := ""

	// Plain text part
	body += fmt.Sprintf("--%s\r\n", boundary)
	body += "Content-Type: text/plain; charset=\"UTF-8\"\r\n"
	body += "Content-Transfer-Encoding: quoted-printable\r\n"
	body += "\r\n"
	body += textBody + "\r\n"

	// HTML part
	body += fmt.Sprintf("--%s\r\n", boundary)
	body += "Content-Type: text/html; charset=\"UTF-8\"\r\n"
	body += "Content-Transfer-Encoding: quoted-printable\r\n"
	body += "\r\n"
	body += htmlBody + "\r\n"

	// Closing boundary
	body += fmt.Sprintf("--%s--\r\n", boundary)

	// 3. รวมร่าง Message
	message := []byte(header + body)

	// 4. ตรวจเช็ค Context ว่าไม่ถูก Cancel หรือ Timeout
	select {
	case <-ctx.Done():
		s.logger.Warn("context cancelled before sending email", "to", to)
		return ctx.Err()
	default:
	}

	// 5. Authentication & Send with timeout
	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Create a channel to send the result
	errChan := make(chan error, 1)
	go func() {
		auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
		errChan <- smtpSendMail("smtp.gmail.com:587", auth, from, []string{to}, message)
	}()

	select {
	case err := <-errChan:
		if err != nil {
			s.logger.Error("Gmail send failed", "error", err, "to", to)
			return err
		}
		s.logger.Info("email sent via Gmail successfully", "to", to)
		return nil
	case <-sendCtx.Done():
		s.logger.Warn("email send timeout", "to", to, "error", sendCtx.Err())
		return sendCtx.Err()
	}
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
