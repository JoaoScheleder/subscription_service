package main

import (
	"bytes"
	"html/template"
	"sync"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  mail.Encryption
	FromAddress string
	FromName    string
	Wait        *sync.WaitGroup
	MailerChan  chan Message
	DoneChan    chan bool
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
	Template    string
}

func (app *Config) ListenForMail() {
	for {
		select {
		case msg := <-app.Mailer.MailerChan:
			if err := app.Mailer.SendMail(msg); err != nil {
				app.ErrorLog.Printf("error sending mail: %v", err)
			}
		case <-app.Mailer.DoneChan:
			return
		}
	}
}

func (m *Mail) SendMail(msg Message) error {

	defer m.Wait.Done()

	if msg.Template == "" {
		msg.Template = "mail"
	}

	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	msg.DataMap = map[string]any{
		"message": msg.Data,
	}

	formattedMsg, err := m.buildHtmlMessage(msg)
	if err != nil {
		return err
	}

	plainMsg, err := m.buildPlainMessage(msg)
	if err != nil {
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.Encryption
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.FromName + "<" + msg.From + ">")
	email.AddTo(msg.To)
	email.SetSubject(msg.Subject)
	email.SetBody(mail.TextPlain, plainMsg)
	email.AddAlternative(mail.TextHTML, formattedMsg)

	for _, attachment := range msg.Attachments {
		email.AddAttachment(attachment)
	}

	if err := email.Send(smtpClient); err != nil {
		return err
	}

	return nil
}

func (m *Mail) buildHtmlMessage(msg Message) (string, error) {
	templateToRender := PATH_TO_TEMPLATES + "/" + msg.Template + ".html.gohtml"
	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (m *Mail) inlineCSS(formattedMsg string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     true,
		CssToAttributes:   true,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(formattedMsg, &options)

	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}

func (m *Mail) buildPlainMessage(msg Message) (string, error) {
	templateToRender := PATH_TO_TEMPLATES + "/" + msg.Template + ".plain.gohtml"
	t, err := template.New("email-html").ParseFiles(templateToRender)

	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}
