package mail

import (
	"fmt"
	"github.com/l4ndm1nes/Weather-API-Application/pkg"
	"go.uber.org/zap"
	"net/smtp"
)

type SMTPMailer struct {
	Host     string
	Port     string
	Username string
	Password string
	From     string
	BaseURL  string
}

func NewSMTPMailer(host, port, username, password, from, baseURL string) *SMTPMailer {
	return &SMTPMailer{
		Host:     host,
		Port:     port,
		Username: username,
		Password: password,
		From:     from,
		BaseURL:  baseURL,
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
	err := smtp.SendMail(addr, auth, m.From, []string{email}, []byte(msg))
	if err != nil {
		pkg.Logger.Error("Failed to send confirmation email",
			zap.String("to", email),
			zap.Error(err),
		)
		return err
	}
	pkg.Logger.Info("Confirmation email sent",
		zap.String("to", email),
	)
	return nil
}

func (m *SMTPMailer) SendWeatherUpdate(email, city, weatherInfo string) error {
	addr := fmt.Sprintf("%s:%s", m.Host, m.Port)
	subject := fmt.Sprintf("Weather update for %s", city)
	body := weatherInfo

	msg := fmt.Sprintf("From: %s\r\nTo: %s\r\nSubject: %s\r\nContent-Type: text/plain; charset=UTF-8\r\n\r\n%s",
		m.From, email, subject, body)
	auth := smtp.PlainAuth("", m.Username, m.Password, m.Host)
	err := smtp.SendMail(addr, auth, m.From, []string{email}, []byte(msg))
	if err != nil {
		pkg.Logger.Error("Failed to send weather update",
			zap.String("to", email),
			zap.String("city", city),
			zap.Error(err),
		)
		return err
	}
	pkg.Logger.Info("Weather update sent",
		zap.String("to", email),
		zap.String("city", city),
	)
	return nil
}
