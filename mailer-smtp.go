package cloudy

import (
	"bytes"
	"context"
	"net/smtp"

	"github.com/appliedres/cloudy/models"
)

var auth smtp.Auth
var addr string

func InitSMTPMailer(MailerConfig *models.Email) {
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
