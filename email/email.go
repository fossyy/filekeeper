package email

import (
	"bytes"
	"context"
	emailView "github.com/fossyy/filekeeper/view/email"
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

func init() {
	mailServer := NewSmtpServer("mail.fossy.my.id", 25, "test@fossy.my.id", "Test123456")
	var buffer bytes.Buffer
	emailView.RegistrationEmail("supri", "https://filekeeper.fossy.my.id/verify/avsihfvasihvf71825185318").Render(context.Background(), &buffer)
	mailServer.Send("bagasaulirizki2@gmail.com", "asfasgfasf", buffer.String())
}

//
//m := gomail.NewMessage()
//m.SetHeader("From", "test@fossy.my.id")
//m.SetHeader("To", "bagasaulirizki2@gmail.com")
//m.SetAddressHeader("Cc", "test@fossy.my.id", "test")
//m.SetHeader("Subject", "Hello!")
//m.SetBody("text/html", buffer.String())
//
//d := gomail.NewDialer("mail.fossy.my.id", 25, "test@fossy.my.id", "Test123456")
