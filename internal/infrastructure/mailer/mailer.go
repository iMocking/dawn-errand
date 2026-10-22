// Package mailer 邮件发送实现：SMTP（生产）与 dev outbox（本地/测试）。
package mailer

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/smtp"
	"time"

	"dawn-errand/internal/domain/system"
	"dawn-errand/internal/infrastructure/ent"
	"dawn-errand/internal/infrastructure/ent/mailoutbox"
	"dawn-errand/internal/pkg/metrics"
)

// ConfigProvider 返回当前生效的 SMTP 配置。
type ConfigProvider func() (*system.SMTPConfig, error)

// Sender 实现 application.MailerPort。
type Sender struct {
	client *ent.Client
	cfg    ConfigProvider
}

// NewSender 构建邮件发送器。
func NewSender(client *ent.Client, cfg ConfigProvider) *Sender {
	return &Sender{client: client, cfg: cfg}
}

// Send 发送邮件：先记录 outbox（审计/失败留痕），SMTP 已启用则实际发送。
func (s *Sender) Send(ctx context.Context, to, subject, html string) error {
	if _, err := s.client.MailOutbox.Create().
		SetToEmail(to).SetSubject(subject).SetBody(html).
		Save(ctx); err != nil {
		return err
	}
	cfg, err := s.cfg()
	if err != nil || cfg == nil || cfg.Host == "" {
		// 未配置 SMTP：仅记录 outbox（dev/测试模式）。
		return nil
	}
	if err := sendSMTP(ctx, cfg, to, subject, html); err != nil {
		return fmt.Errorf("smtp 发送失败: %w", err)
	}
	metrics.EmailsSentTotal.Inc()
	return nil
}

// TestConnection 测试 SMTP 连接。
func (s *Sender) TestConnection(ctx context.Context) error {
	cfg, err := s.cfg()
	if err != nil {
		return err
	}
	if cfg == nil || cfg.Host == "" {
		return fmt.Errorf("SMTP 未配置")
	}
	return testSMTP(ctx, cfg)
}

// TestSMTP 使用给定配置测试 SMTP 连接（配置管理页「测试连接」）。
func TestSMTP(ctx context.Context, c *system.SMTPConfig) error {
	if c == nil || c.Host == "" {
		return fmt.Errorf("SMTP 未配置")
	}
	return testSMTP(ctx, c)
}

func smtpAddr(c *system.SMTPConfig) string {
	if c.Port == 0 {
		c.Port = 587
	}
	return net.JoinHostPort(c.Host, fmt.Sprintf("%d", c.Port))
}

func smtpAuth(c *system.SMTPConfig) smtp.Auth {
	if c.Username == "" {
		return nil
	}
	return smtp.PlainAuth("", c.Username, c.Password, c.Host)
}

func sendSMTP(ctx context.Context, c *system.SMTPConfig, to, subject, html string) error {
	addr := smtpAddr(c)
	msg := fmt.Sprintf("To: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/html; charset=UTF-8\r\n\r\n%s",
		to, subject, html)
	timeout := time.Duration(c.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	if c.Encryption == "ssl" || c.Encryption == "tls" {
		// 隐式 TLS（465）
		tconn := tls.Client(conn, &tls.Config{ServerName: c.Host})
		client, err := smtp.NewClient(tconn, c.Host)
		if err != nil {
			return err
		}
		defer client.Close()
		return sendViaClient(client, c, to, msg)
	}
	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return err
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: c.Host}); err != nil {
			return err
		}
	}
	return sendViaClient(client, c, to, msg)
}

func sendViaClient(client *smtp.Client, c *system.SMTPConfig, to, msg string) error {
	if auth := smtpAuth(c); auth != nil {
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	from := c.FromEmail
	if from == "" {
		from = c.Username
	}
	if err := client.Mail(from); err != nil {
		return err
	}
	if err := client.Rcpt(to); err != nil {
		return err
	}
	w, err := client.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(msg)); err != nil {
		return err
	}
	return w.Close()
}

func testSMTP(ctx context.Context, c *system.SMTPConfig) error {
	addr := smtpAddr(c)
	timeout := time.Duration(c.TimeoutSec) * time.Second
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	d := net.Dialer{Timeout: timeout}
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	if c.Encryption == "ssl" || c.Encryption == "tls" {
		tconn := tls.Client(conn, &tls.Config{ServerName: c.Host})
		client, err := smtp.NewClient(tconn, c.Host)
		if err != nil {
			return err
		}
		defer client.Close()
		return verifyAuth(client, c)
	}
	client, err := smtp.NewClient(conn, c.Host)
	if err != nil {
		return err
	}
	defer client.Close()
	if ok, _ := client.Extension("STARTTLS"); ok {
		if err := client.StartTLS(&tls.Config{ServerName: c.Host}); err != nil {
			return err
		}
	}
	return verifyAuth(client, c)
}

func verifyAuth(client *smtp.Client, c *system.SMTPConfig) error {
	if c.Username == "" {
		return nil
	}
	if ok, _ := client.Extension("AUTH"); !ok {
		return fmt.Errorf("服务器不支持 SMTP AUTH")
	}
	return client.Auth(smtp.PlainAuth("", c.Username, c.Password, c.Host))
}

// OutboxList 查询 dev 邮件发件箱（仅 dev 环境接口使用）。
func (s *Sender) OutboxList(ctx context.Context, limit int) ([]*ent.MailOutbox, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.client.MailOutbox.Query().
		Order(ent.Desc(mailoutbox.FieldCreatedAt)).
		Limit(limit).
		All(ctx)
}
