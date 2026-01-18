package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"passiontree/internal/config"
	"time"

	"github.com/mailersend/mailersend-go"
)

type EmailService interface {
	SendVerificationEmail(to, token string) error
	SendPasswordResetEmail(to, token string) error
}

type emailServiceImpl struct {
	config *config.Config
}

func NewEmailService(cfg *config.Config) EmailService {
	return &emailServiceImpl{
		config: cfg,
	}
}

// SendVerificationEmail sends an email verification link using MailerSend API
func (s *emailServiceImpl) SendVerificationEmail(to, token string) error {
	if s.config.MailerSendAPIKey == "" {
		return fmt.Errorf("MailerSend API key is not set")
	}

	if s.config.SMTPFromEmail == "" {
		return fmt.Errorf("SMTP_FROM_EMAIL is not set; please configure a verified sender address")
	}

	// ใช้ API Key จาก config
	ms := mailersend.NewMailersend(s.config.MailerSendAPIKey)

	// ตั้งค่า timeout context 10 วินาที
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ตั้งค่าข้อมูลอีเมล
	subject := "รหัสยืนยันตัวตน - Passiontree"
	text := fmt.Sprintf("ยินดีต้อนรับสู่ Passiontree!\n\nรหัสยืนยันตัวตนของคุณคือ: %s\n\nกรุณากรอกรหัสนี้ในแอปพลิเคชันเพื่อยืนยันอีเมลของคุณ\n\nรหัสนี้จะหมดอายุใน 15 นาที\n\nหากคุณไม่ได้สร้างบัญชีนี้ กรุณาเพิกเฉยอีเมลนี้", token)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #4CAF50; color: white; padding: 10px; text-align: center; }
        .content { padding: 20px; background-color: #f4f4f4; }
        .code-box { background-color: #fff; border: 2px dashed #4CAF50; padding: 20px; text-align: center; margin: 20px 0; border-radius: 10px; }
        .code { font-size: 36px; font-weight: bold; letter-spacing: 8px; color: #4CAF50; font-family: 'Courier New', monospace; }
        .footer { margin-top: 20px; text-align: center; font-size: 12px; color: #666; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>Passiontree</h1>
        </div>
        <div class="content">
            <h2><b>ยินดีต้อนรับสู่ Passiontree!</b></h2>
            <p>ขอบคุณที่ลงทะเบียนกับ Passiontree!</p>
            <p>กรุณาใช้รหัสยืนยันด้านล่างเพื่อยืนยันอีเมลของคุณ:</p>
            <div class="code-box">
                <div class="code">%s</div>
            </div>
            <p style="text-align: center; color: #666;">
                <small>กรุณากรอกรหัส 6 หลักนี้ในแอปพลิเคชัน</small>
            </p>
            <p><strong>⏰ รหัสนี้จะหมดอายุใน 15 นาที</strong></p>
        </div>
        <div class="footer">
            <p>หากคุณไม่ได้สร้างบัญชีนี้ กรุณาเพิกเฉยอีเมลนี้</p>
            <p>เพื่อความปลอดภัย โปรดอย่าแชร์รหัสนี้ให้ผู้อื่น</p>
        </div>
    </div>
</body>
</html>
`, token)

	// ตั้งค่าผู้ส่ง (ใช้โดเมนที่จดทะเบียนไว้กับ MailerSend)
	from := mailersend.From{
		Name:  "Passiontree Team",
		Email: s.config.SMTPFromEmail, // ต้องเป็นโดเมนที่ verify แล้วใน MailerSend
	}

	// ตั้งค่าผู้รับ
	recipients := []mailersend.Recipient{
		{
			Name:  "User",
			Email: to,
		},
	}

	// สร้างและส่งข้อความ
	message := ms.Email.NewMessage()

	message.SetFrom(from)
	message.SetRecipients(recipients)
	message.SetSubject(subject)
	message.SetHTML(html)
	message.SetText(text)

	res, err := ms.Email.Send(ctx, message)
	if err != nil {
		fmt.Printf("Full Error: %+v\n", err)
		return fmt.Errorf("failed to send email via MailerSend: %w", err)
	}

	// แสดงข้อมูลการส่งสำเร็จ
	messageID := res.Header.Get("X-Message-Id")
	fmt.Printf("Email sent successfully to %s. Message ID: %s\n", to, messageID)
	fmt.Printf("MailerSend response: %+v\n", res)

	return nil
}

// SendPasswordResetEmail sends a password reset code email
func (s *emailServiceImpl) SendPasswordResetEmail(to, token string) error {
	if s.config.MailerSendAPIKey == "" {
		return fmt.Errorf("MailerSend API key is not set")
	}

	if s.config.SMTPFromEmail == "" {
		return fmt.Errorf("SMTP_FROM_EMAIL is not set; please configure a verified sender address")
	}

	// ใช้ API Key จาก config
	ms := mailersend.NewMailersend(s.config.MailerSendAPIKey)

	// ตั้งค่า timeout context 10 วินาที
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ตั้งค่าข้อมูลอีเมล
	subject := "รีเซ็ตรหัสผ่าน - Passiontree"
	text := fmt.Sprintf("คำขอรีเซ็ตรหัสผ่าน\n\nรหัสรีเซ็ตของคุณคือ: %s\n\nกรุณากรอกรหัสนี้ในแอปพลิเคชันเพื่อตั้งรหัสผ่านใหม่\n\nรหัสนี้จะหมดอายุใน 15 นาที\n\nหากคุณไม่ได้ขอรีเซ็ตรหัสผ่าน กรุณาเพิกเฉยอีเมลนี้", token)

	html := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <style>
        body { font-family: Arial, sans-serif; line-height: 1.6; }
        .container { max-width: 600px; margin: 0 auto; padding: 20px; }
        .header { background-color: #FF5722; color: white; padding: 10px; text-align: center; }
        .content { padding: 20px; background-color: #f4f4f4; }
        .code-box { background-color: #fff; border: 2px dashed #FF5722; padding: 20px; text-align: center; margin: 20px 0; border-radius: 10px; }
        .code { font-size: 36px; font-weight: bold; letter-spacing: 8px; color: #FF5722; font-family: 'Courier New', monospace; }
        .footer { margin-top: 20px; text-align: center; font-size: 12px; color: #666; }
        .warning { background-color: #fff3cd; border-left: 4px solid #ff9800; padding: 10px; margin: 10px 0; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>🔒 Passiontree</h1>
        </div>
        <div class="content">
            <h2><b>รีเซ็ตรหัสผ่าน</b></h2>
            <p>คุณได้ขอรีเซ็ตรหัสผ่านสำหรับบัญชี Passiontree ของคุณ</p>
            <p>กรุณาใช้รหัสด้านล่างเพื่อตั้งรหัสผ่านใหม่:</p>
            <div class="code-box">
                <div class="code">%s</div>
            </div>
            <p style="text-align: center; color: #666;">
                <small>กรุณากรอกรหัส 6 หลักนี้ในแอปพลิเคชัน</small>
            </p>
            <div class="warning">
                <strong>⚠️ สำคัช:</strong> รหัสนี้จะหมดอายุใน 15 นาที
            </div>
        </div>
        <div class="footer">
            <p>หากคุณไม่ได้ขอรีเซ็ตรหัสผ่าน กรุณาเพิกเฉยอีเมลนี้</p>
            <p>เพื่อความปลอดภัย โปรดอย่าแชร์รหัสนี้ให้ผู้อื่น</p>
        </div>
    </div>
</body>
</html>
`, token)

	// ตั้งค่าผู้ส่ง
	from := mailersend.From{
		Name:  "Passiontree Security",
		Email: s.config.SMTPFromEmail,
	}

	// ตั้งค่าผู้รับ
	recipients := []mailersend.Recipient{
		{
			Name:  "User",
			Email: to,
		},
	}

	// สร้างและส่งข้อความ
	message := ms.Email.NewMessage()

	message.SetFrom(from)
	message.SetRecipients(recipients)
	message.SetSubject(subject)
	message.SetHTML(html)
	message.SetText(text)

	res, err := ms.Email.Send(ctx, message)
	if err != nil {
		fmt.Printf("Full Error: %+v\n", err)
		return fmt.Errorf("failed to send password reset email via MailerSend: %w", err)
	}

	// แสดงข้อมูลการส่งสำเร็จ
	messageID := res.Header.Get("X-Message-Id")
	fmt.Printf("Password reset email sent successfully to %s. Message ID: %s\n", to, messageID)

	return nil
}

// GenerateVerificationToken generates a random 6-digit code for email verification
func GenerateVerificationToken() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	// Generate 6-digit code (100000-999999)
	code := (int(bytes[0])<<24|int(bytes[1])<<16|int(bytes[2])<<8|int(bytes[3]))%900000 + 100000
	return fmt.Sprintf("%06d", code), nil
}

// GetVerificationTokenExpiry returns expiry time (15 minutes from now for code-based verification)
func GetVerificationTokenExpiry() time.Time {
	return time.Now().Add(15 * time.Minute)
}
