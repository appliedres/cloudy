package cloudy

import (
	"bytes"
	"context"
	"net"
	"net/smtp"
	"time"

	"github.com/appliedres/cloudy/models"
)

var auth smtp.Auth
var addr string

func InitSMTPMailer(ctx context.Context, MailerConfig *models.Email) {
	auth = smtp.PlainAuth("", MailerConfig.From, MailerConfig.Password, MailerConfig.Host)
	addr = MailerConfig.Host + ":" + MailerConfig.Port
}

func SendSMTPMail(ctx context.Context, to []string, from string, body bytes.Buffer) error {
	err := smtp.SendMail(addr, auth, from, to, body.Bytes())
	if err != nil {
		return err
	}

	return nil
}

func SendSMTPMailNoAuth(ctx context.Context, server string, to []string, from string, body bytes.Buffer) error {

	//verify connectivity as smtp.Dial blocks
	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.Dial("tcp", server)
	if err != nil {
		return err
	}
	conn.Close()

	client, err := smtp.Dial(server)
	if err != nil {
		return err
	}

	defer client.Quit()
	defer client.Close()

	err = client.Mail(from)
	if err != nil {
		return err
	}

	for i := range to {
		err = client.Rcpt(to[i])
		if err != nil {
			return err
		}
	}

	writer, err := client.Data()
	if err != nil {
		return err
	}
	defer writer.Close()

	_, err = writer.Write(body.Bytes())
	if err != nil {
		return err
	}

	return nil
}
