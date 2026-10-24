// Package mail provides OTP email sending. ConsoleSender for dev, SMTPSender for prod.
package mail

import (
	"fmt"
	"log"
	"net/smtp"
)

// Sender sends OTP codes via email.
type Sender interface {
	SendOTP(to, code string) error
}

// ConsoleSender prints the OTP to stdout (development mode).
type ConsoleSender struct{}

func (ConsoleSender) SendOTP(to, code string) error {
	log.Printf("[DEV] OTP for %s: %s", to, code)
	return nil
}

// SMTPSender sends real emails via SMTP.
type SMTPSender struct {
	Host     string
	Port     string
	User     string
	Pass     string
	MailFrom string
}

func (s *SMTPSender) SendOTP(to, code string) error {
	addr := s.Host + ":" + s.Port
	auth := smtp.PlainAuth("", s.User, s.Pass, s.Host)
	subject := "Your Smart Mess verification code"
	body := fmt.Sprintf("Your verification code is: %s\n\nThis code expires in 10 minutes. Do not share it with anyone.", code)
	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s",
		s.MailFrom, to, subject, body)
	return smtp.SendMail(addr, auth, s.MailFrom, []string{to}, []byte(msg))
}
