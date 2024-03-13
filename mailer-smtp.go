package cloudy

import (
	"bytes"
	"context"
	"net"
	"net/smtp"
	"time"

	"github.com/appliedres/cloudy/models"
)

var authCfg smtp.Auth
var addrCfg string
var fromCfg string

func SendSMTPMail(ctx context.Context, mailerConfig *models.Email, to []string, body bytes.Buffer) error {
	var err error
	initSMTPMailer(mailerConfig)

	dialer := net.Dialer{Timeout: 2 * time.Second}
	conn, err := dialer.Dial("tcp", addrCfg)
	if err != nil {
		return err
	}
	conn.Close()

	if mailerConfig.AuthenticationRequired {
		err = sendSMTPMailAuth(addrCfg, authCfg, to, fromCfg, body)
	} else {
		err = SendSMTPMailNoAuth(ctx, addrCfg, to, fromCfg, body)
	}
	return err
}

func initSMTPMailer(mailerConfig *models.Email) {
	authCfg = smtp.PlainAuth("", mailerConfig.From, mailerConfig.Password, mailerConfig.Host)
	addrCfg = mailerConfig.Host + ":" + mailerConfig.Port
	fromCfg = mailerConfig.From
}

func sendSMTPMailAuth(server string, auth smtp.Auth, to []string, from string, body bytes.Buffer) error {
	err := smtp.SendMail(server, auth, from, to, body.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// change to private after user-api uses SendSMTPMail with authreq parameter
func SendSMTPMailNoAuth(ctx context.Context, server string, to []string, from string, body bytes.Buffer) error {

	//verify connectivity as smtp.Dial blocks
	// remove dialer block after user-api uses SendSMTPMail with authreq parameter
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
