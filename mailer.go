package cloudy

import (
	"context"
	"errors"
	"io"
	"strings"
	"text/template"

	"gopkg.in/gomail.v2"
)

type Attachment struct {
	Name   string
	Path   string
	Data   []byte
	Reader io.Reader
}

type Email struct {
	To      []string
	CC      []string
	BCC     []string
	From    string
	Body    string
	HTML    bool
	Subject string

	Attachments []Attachment
}

var DefaultEmailer *Mailer
var EmailerProviders = NewProviderRegistry[Emailer]()

type Emailer interface {
	Send(ctx context.Context, message *gomail.Message) error
}

type Mailer struct {
	Before  []BeforeInterceptor[Email]
	After   []AfterInterceptor[Email]
	Emailer Emailer
}

func NewMailer(ctx context.Context, provider string, cfg interface{}) {

}

func (m *Mailer) ToGoMailMessage(ctx context.Context, message *Email) *gomail.Message {
	msg := gomail.NewMessage()
	msg.SetHeader("From", message.From)
	msg.SetHeader("To", message.To...)
	msg.SetHeader("Subject", message.Subject)
	if message.HTML {
		msg.SetBody("text/html", message.Body)
	} else {
		msg.SetBody("text/text", message.Body)

	}

	for _, a := range message.Attachments {
		if a.Path != "" {
			msg.Attach(a.Path)
		} else if a.Data != nil {
			msg.Attach(a.Name, gomail.SetCopyFunc(func(w io.Writer) error {
				_, err := w.Write(a.Data)
				return err
			}))
		}
	}
	return msg
}

func (m *Mailer) Send(ctx context.Context, item *Email) (err error) {
	for _, before := range m.Before {
		item, err = before.BeforeAction(ctx, item)
		if err != nil {
			return
		}
	}

	msg := m.ToGoMailMessage(ctx, item)

	err = m.Emailer.Send(ctx, msg)
	if err != nil {
		return
	}

	for _, before := range m.After {
		item, err = before.AfterAction(ctx, item)
		if err != nil {
			return
		}
	}

	return
}

type TemplatedEmail struct {
	Email
	BodyTemplate    *template.Template
	SubjectTemplate *template.Template
	InlineStyles    bool
}

func (e *Email) Send(ctx context.Context) error {
	if DefaultEmailer == nil {
		return errors.New("no default emailer")
	}

	return DefaultEmailer.Send(ctx, e)
}

func (t *TemplatedEmail) Render(ctx context.Context, data interface{}) (email *Email, err error) {
	body, subject, err := t.RenderRaw(ctx, data)
	if err != nil {
		return nil, err
	}

	email = &Email{
		To:          t.To,
		CC:          t.CC,
		BCC:         t.BCC,
		From:        t.From,
		Body:        body,
		HTML:        t.HTML,
		Subject:     subject,
		Attachments: t.Attachments,
	}

	return
}

func (t *TemplatedEmail) Send(ctx context.Context, data interface{}) error {
	email, err := t.Render(ctx, data)
	if err != nil {
		return err
	}
	return email.Send(ctx)
}

func (t *TemplatedEmail) RenderRaw(ctx context.Context, data interface{}) (body string, subject string, err error) {
	if t.BodyTemplate != nil {
		var bodyStr strings.Builder
		err := t.BodyTemplate.Execute(&bodyStr, data)
		if err != nil {
			return "", "", err
		}
		body = bodyStr.String()
	}

	if t.SubjectTemplate != nil {
		var subStr strings.Builder
		err := t.SubjectTemplate.Execute(&subStr, data)
		if err != nil {
			return "", "", err
		}
		subject = subStr.String()
	}
	return
}

// type MessageData struct {
// 	Field string
// }
// func init() {
// 	DefaultEmailer = &Mailer{
// 		Before: []BeforeInterceptor[Email] {
// 			NewMandatoryBCC()
// 		},
// 		Emailer: NewSESEmailer(cfg),
// 		After: []BeforeInterceptor[Email] {
// 			NewEmailLogger()
// 		}
// 	}
// }
// func idea() {
// 	t := &TemplatedEmail{
// 		Email: Email{
// 			From: "system@company.com",
// 			To:   []string{"admin@company.com"},
// 		},
// 		BodyTemplate:    template.Must(template.New("testbody").Parse(`This is My Body {{ .Field }}`)),
// 		SubjectTemplate: template.Must(template.New("testbody").Parse(`This is My Subject {{ .Field }}`)),
// 	}

// 	data := MessageData{
// 		Field: "Foo",
// 	}

// 	err := t.Send(context.Background(), data)
// 	if err != nil {
// 		panic(err)
// 	}

// 	/// OR

// 	t.To = []string{"abc@abc.com"}
// }
