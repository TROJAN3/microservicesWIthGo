package main

import (
	"bytes"
	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
	"html/template"
	_ "io"
	"log"
	"time"
)

type Mail struct {
	Domain      string `json:"domain"`
	Host        string `json:"host"`
	Port        int    `json:"port"`
	Username    string `json:"username"`
	Password    string `json:"password"`
	Encryption  string `json:"encryption"`
	FromAddress string `json:"from"`
	FromName    string `json:"from_name"`
}

type Message struct {
	From        string   `json:"from"`
	FromName    string   `json:"from_name"`
	To          string   `json:"to"`
	Subject     string   `json:"subject"`
	Attachments []string `json:"attachments"`
	Data        any
	DataMap     map[string]any
}

func (m *Mail) SendSMTPMessage(msg Message) error {
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

	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		log.Println("error building html message", err)
		return err
	}

	plainMessage, err := m.buildPlainMessage(msg)
	if err != nil {
		log.Println("error building plain message", err)
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtClient, err := server.Connect()
	if err != nil {
		log.Println("error connecting to smtp server", err)
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).AddTo(msg.To).SetSubject(msg.Subject)
	email.SetBody(mail.TextPlain, plainMessage)
	email.AddAlternative(mail.TextHTML, formattedMessage)

	if len(msg.Attachments) > 0 {
		for _, attachment := range msg.Attachments {
			email.AddAttachment(attachment)
		}
	}

	err = email.Send(smtClient)
	if err != nil {
		log.Println("error sending email", err)
		return err
	}

	return nil
}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := "./templates/mail.html.gohtml"

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		log.Println("error parsing email template", err)
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		log.Println("error executing email template", err)
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		log.Println("error inlining css", err)
		return "", err
	}

	return formattedMessage, nil

}

func (m *Mail) buildPlainMessage(msg Message) (string, error) {
	templateToRender := "./templates/mail.plain.gohtml"

	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		log.Println("error parsing plain email template", err)
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.DataMap); err != nil {
		log.Println("error executing plain email template", err)
		return "", err
	}

	plainMessage := tpl.String()
	return plainMessage, nil

}

func (m *Mail) inlineCSS(s string) (string, error) {

	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   true,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)

	html, err := prem.Transform()
	if err != nil {
		log.Println("error inlining css", err)
		return "", err
	}

	return html, nil
}

func (m *Mail) getEncryption(e string) mail.Encryption {
	switch e {
	case "SSL":

		return mail.EncryptionSSLTLS
	case "TLS":
		return mail.EncryptionSTARTTLS
	case "none", "":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
