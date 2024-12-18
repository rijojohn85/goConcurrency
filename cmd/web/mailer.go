package main

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"
	"time"

	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

type Mail struct {
	Wait        *sync.WaitGroup
	MailerChan  chan Message
	ErrorChan   chan error
	DoneChan    chan bool
	Domain      string
	Host        string
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	Port        int
}

type Message struct {
	Data        any
	DataMap     map[string]any
	From        string
	FromName    string
	Subject     string
	To          string
	Template    string
	Attachments []string
}

// a function to listen for messages on the MailerChan
func (app *Config) listenForMail() {
	for {
		select {
		case msg := <-app.Mailer.MailerChan:
			go app.Mailer.sendMail(msg, app.Mailer.ErrorChan)
		case err := <-app.Mailer.ErrorChan:
			app.ErrorLog.Println("Mailer error:", err)
		case <-app.Mailer.DoneChan:
			return
		}
	}
}

func (m *Mail) sendMail(msg Message, errorChan chan error) {
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
	data := map[string]any{
		"message": msg.Data,
	}
	msg.DataMap = data
	// build html mail
	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		errorChan <- err
	}
	plainMessage, err := m.buildPlainMessage(msg)
	if err != nil {
		errorChan <- err
	}
	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = time.Second * 10
	server.SendTimeout = time.Second * 10

	smtpClient, err := server.Connect()
	if err != nil {
		errorChan <- err
	}
	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)
	email.SetBody(mail.TextPlain, plainMessage)
	email.AddAlternative(mail.TextHTML, formattedMessage)
	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}
	err = email.Send(smtpClient)
	if err != nil {
		errorChan <- err
	}
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("./cmd/web/templates/%s.html.gohtml", msg.Template)
	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		return "", err
	}
	formattedMessage := tpl.String()

	formattedMessage, err = m.inLineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (m *Mail) buildPlainMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("./cmd/web/templates/%s.plain.gohtml", msg.Template)
	t, err := template.New("email-plain").ParseFiles(templateToRender)
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

func (m *Mail) inLineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}
	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}
	html, err := prem.Transform()
	if err != nil {
		return "", err
	}
	return html, nil
}

func (m *Mail) getEncryption(e string) mail.Encryption {
	switch e {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
