package mail

import (
	"fmt"
	"net/smtp"
	"os"
)

type SMTPMailer struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	BaseURL  string
}

func NewSMTPMailerFromEnv() *SMTPMailer {
	return &SMTPMailer{
		Host:     os.Getenv("SMTP_HOST"),
		Port:     os.Getenv("SMTP_PORT"),
		Username: os.Getenv("SMTP_USER"),
		Password: os.Getenv("SMTP_PASS"),
		From:     os.Getenv("SMTP_FROM"),
		BaseURL:  os.Getenv("BASE_URL"),
	}
}

func (m *SMTPMailer) SendConfirmation(email, token string) error {
	addr := fmt.Sprintf("%s:%s", m.Host, m.Port)
	subject := "Confirm your weather subscription"
	link := fmt.Sprintf("%s/api/confirm/%s", m.BaseURL, token)
	body := fmt.Sprintf("To confirm your subscription, click the link: %s", link)

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.From, email, subject, body)
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
	return smtp.SendMail(addr, auth, m.From, []string{email}, []byte(msg))
}

func (m *SMTPMailer) SendWeatherUpdate(email, city, weatherInfo string) error {
	addr := fmt.Sprintf("%s:%s", m.Host, m.Port)
	subject := fmt.Sprintf("Weather update for %s", city)
	body := weatherInfo

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.From, email, subject, body)
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
	return smtp.SendMail(addr, auth, m.From, []string{email}, []byte(msg))
}
