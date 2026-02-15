package service

import (
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
	html := buildVerificationEmailHTML(token)
	
	return s.sendEmail(to, subject, html, text, "Passiontree Team")
}

func (s *emailServiceImpl) SendPasswordResetEmail(to, token string) error {
	subject := "รีเซ็ตรหัสผ่าน - Passiontree"
	text := fmt.Sprintf("รหัสรีเซ็ตรหัสผ่าน: %s\n\nรหัสนี้จะหมดอายุใน 15 นาที", token)
	html := buildPasswordResetEmailHTML(token)
	
	return s.sendEmail(to, subject, html, text, "Passiontree Security")
}

func (s *emailServiceImpl) SendSecurityAlertEmail(to, userID string) error {
	subject := "แจ้งเตือนความปลอดภัย - Passiontree"
	now := time.Now().Format("02/01/2006 15:04:05")
	text := fmt.Sprintf("ตรวจพบกิจกรรมที่น่าสงสัยในบัญชีของคุณ\n\nUser ID: %s\nเวลา: %s\n\nเราได้ล็อกเอาท์ทุกอุปกรณ์เพื่อความปลอดภัย\nกรุณาเปลี่ยนรหัสผ่านทันที", userID, now)
	html := buildSecurityAlertEmailHTML(userID, now)
	
	return s.sendEmail(to, subject, html, text, "Passiontree Security Team")
}

func (s *emailServiceImpl) sendEmail(to, subject, html, text, fromName string) error {
	if s.config.MailerSendAPIKey == "" || s.config.SMTPFromEmail == "" {
		return apperror.NewInternal("email service not configured")
	}

	ms := mailersend.NewMailersend(s.config.MailerSendAPIKey)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	message := ms.Email.NewMessage()
	message.SetFrom(mailersend.From{Name: fromName, Email: s.config.SMTPFromEmail})
	message.SetRecipients([]mailersend.Recipient{{Name: "User", Email: to}})
	message.SetSubject(subject)
	message.SetHTML(html)
	message.SetText(text)

	if _, err := ms.Email.Send(ctx, message); err != nil {
		s.logger.Error("send email failed", "to", to, "error", err)
		return apperror.NewInternal("failed to send email: %w", err)
	}

	s.logger.Info("email sent", "to", to, "subject", subject)
	return nil
}

func buildVerificationEmailHTML(token string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8">
<style>
body{font-family:Arial,sans-serif;line-height:1.6}
.container{max-width:600px;margin:0 auto;padding:20px}
.header{background:#4CAF50;color:#fff;padding:10px;text-align:center}
.content{padding:20px;background:#f4f4f4}
.code-box{background:#fff;border:2px dashed #4CAF50;padding:20px;text-align:center;margin:20px 0;border-radius:10px}
.code{font-size:36px;font-weight:bold;letter-spacing:8px;color:#4CAF50;font-family:monospace}
.footer{margin-top:20px;text-align:center;font-size:12px;color:#666}
</style>
</head>
<body>
<div class="container">
<div class="header"><h1>Passiontree</h1></div>
<div class="content">
<h2>ยินดีต้อนรับสู่ Passiontree!</h2>
<p>กรุณาใช้รหัสยืนยันด้านล่าง:</p>
<div class="code-box"><div class="code">%s</div></div>
<p style="text-align:center;color:#666"><small>รหัส 6 หลัก</small></p>
<p><strong>หมดอายุใน 15 นาที</strong></p>
</div>
<div class="footer">
<p>หากคุณไม่ได้สร้างบัญชีนี้ กรุณาเพิกเฉยอีเมลนี้</p>
</div>
</div>
</body>
</html>`, token)
}

func buildPasswordResetEmailHTML(token string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8">
<style>
body{font-family:Arial,sans-serif;line-height:1.6}
.container{max-width:600px;margin:0 auto;padding:20px}
.header{background:#FF5722;color:#fff;padding:10px;text-align:center}
.content{padding:20px;background:#f4f4f4}
.code-box{background:#fff;border:2px dashed #FF5722;padding:20px;text-align:center;margin:20px 0;border-radius:10px}
.code{font-size:36px;font-weight:bold;letter-spacing:8px;color:#FF5722;font-family:monospace}
.warning{background:#fff3cd;border-left:4px solid #ff9800;padding:10px;margin:10px 0}
.footer{margin-top:20px;text-align:center;font-size:12px;color:#666}
</style>
</head>
<body>
<div class="container">
<div class="header"><h1>Passiontree</h1></div>
<div class="content">
<h2>รีเซ็ตรหัสผ่าน</h2>
<p>กรุณาใช้รหัสด้านล่าง:</p>
<div class="code-box"><div class="code">%s</div></div>
<div class="warning"><strong>รหัสหมดอายุภายใน 15 นาที</strong></div>
</div>
<div class="footer">
<p>หากคุณไม่ได้ขอรีเซ็ตรหัสผ่าน กรุณาเพิกเฉยอีเมลนี้</p>
</div>
</div>
</body>
</html>`, token)
}

func buildSecurityAlertEmailHTML(userID, timestamp string) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head><meta charset="UTF-8">
<style>
body{font-family:Arial,sans-serif;line-height:1.6}
.container{max-width:600px;margin:0 auto;padding:20px}
.header{background:#d32f2f;color:#fff;padding:20px;text-align:center}
.content{padding:20px;background:#f4f4f4}
.alert{background:#ffebee;border-left:5px solid #d32f2f;padding:15px;margin:20px 0}
.action{background:#fff;border:2px solid #ff9800;padding:20px;margin:20px 0;border-radius:10px}
.item{margin:10px 0;padding-left:20px}
.footer{margin-top:20px;text-align:center;font-size:12px;color:#666}
</style>
</head>
<body>
<div class="container">
<div class="header"><h1> แจ้งเตือนความปลอดภัย</h1></div>
<div class="content">
<div class="alert">
<h3 style="color:#d32f2f;margin-top:0">ตรวจพบกิจกรรมที่น่าสงสัย</h3>
<p>มีการพยายามใช้ Token ที่ผิดปกติ อาจเป็นการโจรกรรม</p>
</div>
<div class="action">
<h4 style="color:#ff9800">การดำเนินการ</h4>
<div class="item">ล็อกเอาท์ทุกอุปกรณ์</div>
<div class="item">เพิกถอน Token ทั้งหมด</div>
</div>
<div class="action">
<h4 style="color:#d32f2f">สิ่งที่คุณควรทำ</h4>
<div class="item">1. เข้าสู่ระบบอีกครั้ง</div>
<div class="item">2. <strong>เปลี่ยนรหัสผ่านทันที</strong></div>
<div class="item">3. ตรวจสอบอุปกรณ์ที่เคยเข้าใช้</div>
</div>
<p><strong>รายละเอียด:</strong><br>
User ID: <code>%s</code><br>
เวลา: %s</p>
</div>
<div class="footer">
<p style="color:#d32f2f;font-weight:bold">หากไม่ใช่คุณ กรุณาติดต่อทีมสนับสนุนทันที</p>
</div>
</div>
</body>
</html>`, userID, timestamp)
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
