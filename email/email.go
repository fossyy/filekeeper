package email

import (
	"gopkg.in/gomail.v2"
)

type SmtpServer struct {
	Host     string
	Port     int
	User     string
	Password string
}

type Email interface {
	Send()
}

func NewSmtpServer(Host string, Port int, User string, Password string) *SmtpServer {
	return &SmtpServer{
		Host:     Host,
		Port:     Port,
		User:     User,
		Password: Password,
	}
}

func (mail *SmtpServer) Send(dst string, subject string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", mail.User)
	m.SetHeader("To", dst)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)
	d := gomail.NewDialer(mail.Host, mail.Port, mail.User, mail.Password)

	if err := d.DialAndSend(m); err != nil {
		return err
	}
	return nil
}
