package service

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"mime"
	"mime/quotedprintable"
	"net/mail"
	"net/smtp"
	"strings"
	"time"

	"passiontree/internal/pkg/apperror"
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

func normalizeBodyLineEndings(value string) string {
	value = strings.ReplaceAll(value, "\r\n", "\n")
	value = strings.ReplaceAll(value, "\r", "\n")
	return strings.ReplaceAll(value, "\n", "\r\n")
}

func encodeQuotedPrintableBody(value string) (string, error) {
	var buf bytes.Buffer
	writer := quotedprintable.NewWriter(&buf)
	if _, err := writer.Write([]byte(normalizeBodyLineEndings(value))); err != nil {
		_ = writer.Close()
		return "", err
	}
	if err := writer.Close(); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func (s *emailServiceImpl) SendVerificationEmail(ctx context.Context, to, token string) error {
	subject := "รหัสยืนยันตัวตน - Passion-Tree"
	text := fmt.Sprintf("ยินดีต้อนรับสู่ Passion-Tree!\n\nรหัสยืนยันตัวตนของคุณคือ: %s\n\nรหัสนี้จะหมดอายุใน 5 นาที", token)

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
	text := fmt.Sprintf("รหัสรีเซ็ตรหัสผ่าน: %s\n\nรหัสนี้จะหมดอายุใน 5 นาที", token)

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

	var htmlBuf bytes.Buffer
	data := map[string]string{
		"Subject":  subject,
		"Headline": headline,
		"Message":  message,
	}
	if err := s.templates.notification.Execute(&htmlBuf, data); err != nil {
		s.logger.Error("failed to execute notification template", "error", err)
		return apperror.NewInternal("failed to generate email content")
	}

	return s.sendEmail(ctx, to, subject, htmlBuf.String(), text, "Passiontree Notifications")
}

func (s *emailServiceImpl) sendEmail(ctx context.Context, to, subject, html, text, fromName string) error {
	// ส่งด้วย Gmail SMTP
	return s.sendViaGmail(ctx, to, subject, html, text, fromName)
}

func (s *emailServiceImpl) sendViaGmail(ctx context.Context, to, subject, htmlBody, textBody, fromName string) error {
	from := strings.TrimSpace(s.config.GmailEmail)
	password := strings.ReplaceAll(strings.TrimSpace(s.config.GmailAppPassword), " ", "")

	if from == "" || password == "" {
		return fmt.Errorf("gmail credentials not configured")
	}

	// Sanitize & Encode
	to = sanitizeHeaderValue(to)
	subject = sanitizeHeaderValue(subject)
	fromAddress := mail.Address{Name: fromName, Address: from}
	encodedSubject := mime.QEncoding.Encode("UTF-8", subject)

	encodedTextBody, err := encodeQuotedPrintableBody(textBody)
	if err != nil {
		return fmt.Errorf("failed to encode text body: %w", err)
	}
	encodedHTMLBody, err := encodeQuotedPrintableBody(htmlBody)
	if err != nil {
		return fmt.Errorf("failed to encode html body: %w", err)
	}

	// Generate boundary
	boundaryBuffer := make([]byte, 16)
	if _, err := rand.Read(boundaryBuffer); err != nil {
		return fmt.Errorf("failed to generate boundary: %w", err)
	}
	boundary := fmt.Sprintf("%x", boundaryBuffer)

	// --- 1. สร้าง Message ด้วย bytes.Buffer เพื่อคุมบรรทัดให้แม่นยำ ---
	var msg bytes.Buffer

	// Headers
	msg.WriteString(fmt.Sprintf("From: %s\r\n", fromAddress.String()))
	msg.WriteString(fmt.Sprintf("To: %s\r\n", to))
	msg.WriteString(fmt.Sprintf("Subject: %s\r\n", encodedSubject))
	msg.WriteString("MIME-Version: 1.0\r\n")
	msg.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=\"%s\"\r\n", boundary))
	msg.WriteString("\r\n") // บรรทัดว่างคั่น Header และ Body

	// --- 2. สร้าง Body Parts ---
	// Plain text part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/plain; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(encodedTextBody)
	msg.WriteString("\r\n")

	// HTML part
	msg.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	msg.WriteString("Content-Type: text/html; charset=\"UTF-8\"\r\n")
	msg.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	msg.WriteString("\r\n")
	msg.WriteString(encodedHTMLBody)
	msg.WriteString("\r\n")

	// Closing boundary
	msg.WriteString(fmt.Sprintf("--%s--\r\n", boundary))

	// --- 3. การส่งเมล ---
	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	errChan := make(chan error, 1)
	go func() {
		auth := smtp.PlainAuth("", from, password, "smtp.gmail.com")
		errChan <- smtpSendMail("smtp.gmail.com:587", auth, from, []string{to}, msg.Bytes())
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
		return fmt.Errorf("email send timeout after 15s")
	}
}

// MockOTPCode is a fixed OTP used when SMTP is unavailable (e.g. Render blocks
// outbound SMTP). The login flow uses this instead of GenerateVerificationToken
// so users can verify with a known code without a real email being sent.
const MockOTPCode = "729384"

func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	code := (int(bytes[0])<<24|int(bytes[1])<<16|int(bytes[2])<<8|int(bytes[3]))%900000 + 100000
	return fmt.Sprintf("%06d", code), nil
}

func GetVerificationTokenExpiry() time.Time {
	return time.Now().Add(5 * time.Minute)
}
