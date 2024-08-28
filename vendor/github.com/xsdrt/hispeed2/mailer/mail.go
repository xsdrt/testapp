package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"time"

	apimail "github.com/ainsleyclark/go-mail"    // Allow to use MailGun: SparkPost or SendGrid...
	"github.com/vanng822/go-premailer/premailer" // Used to inline the CSS...
	mail "github.com/xhit/go-simple-mail/v2"     //go-simple-mail; easy way to send mail with a client...
)

// Mail holds the information needed to connect to a SMTP server
type Mail struct {
	Domain      string
	Templates   string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	Jobs        chan Message
	Results     chan Result
	API         string
	APIKey      string
	APIUrl      string
}

// Message is the type for a email message...
type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Template    string
	Attachments []string
	Data        interface{}
}

// Result contains info regarding the status of the sent email message...
type Result struct {
	Success bool
	Error   error
}

// ListenForMail listens to the mail channel and sends mail
// when it recieves a payload.  It runs continually in the background,
// and sends error/sucess messages back on the results channel.
// Note that if an api and api key are set, it will prefer using
// an api to send mail...

func (m *Mail) ListenForMail() {
	for {
		msg := <-m.Jobs    // This chan will listen for incoming message...
		err := m.Send(msg) // Then attempt to send the message...
		if err != nil {
			m.Results <- Result{false, err} // If an error, return the error...
		} else {
			m.Results <- Result{true, nil} // Otherwise success...
		}
	}
}

func (m *Mail) Send(msg Message) error {
	// Use an API or SMTP?
	if len(m.API) > 0 && len(m.APIKey) > 0 && len(m.APIUrl) > 0 && m.API != "smtp" {
		m.ChooseAPI(msg)
	}
	return m.SendSTMPMessage(msg)
}

func (m *Mail) ChooseAPI(msg Message) error {
	switch m.API {
	case "mailgun", "sparkpost", "sendgrid": // This is set in the .env file...
		return m.SendUsingAPI(msg, m.API)
	default:
		return fmt.Errorf("unknown api %s; only mailgun, sparkpost or sendgrid supported", m.API)
	}
}

func (m *Mail) SendUsingAPI(msg Message, transport string) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	cfg := apimail.Config{
		URL:         m.APIUrl,
		APIKey:      m.APIKey,
		Domain:      m.Domain,
		FromAddress: msg.From,
		FromName:    msg.FromName,
	}

	driver, err := apimail.NewClient(transport, cfg)
	if err != nil {
		return err
	}

	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	tx := &apimail.Transmission{
		Recipients: []string{msg.To},
		Subject:    msg.Subject,
		HTML:       formattedMessage,
		PlainText:  plainMessage,
	}

	// add attachments
	err = m.addAPIAttachments(msg, tx)
	if err != nil {
		return err
	}

	_, err = driver.Send(tx)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mail) addAPIAttachments(msg Message, tx *apimail.Transmission) error {
	if len(msg.Attachments) > 0 {
		var attachments []apimail.Attachment

		for _, x := range msg.Attachments {
			var attach apimail.Attachment
			content, err := os.ReadFile(x)
			if err != nil {
				return err
			}

			fileName := filepath.Base(x)
			attach.Bytes = content
			attach.Filename = fileName
			attachments = append(attachments, attach)
		}

		tx.Attachments = attachments
	}

	return nil
}

// SendSTMPMessage builds and sends an email using STMP; This is called by ListenForMail,
// an can aslo be called directly when wanted/necessary...
func (m *Mail) SendSTMPMessage(msg Message) error {
	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
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

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject)

	email.SetBody(mail.TextHTML, formattedMessage)
	email.AddAlternative(mail.TextPlain, plainMessage)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		return err
	}

	return nil
}

func (m *Mail) getEncryption(e string) mail.Encryption {
	switch e {
	case "tls": // Usually primary in production...
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSL
	case "none": // Only during develpment...
		return mail.EncryptionNone
	default: // If the user does not specify the encrption assume want to use tls by default
		return mail.EncryptionSTARTTLS
	}

}

func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.html.tmpl", m.Templates, msg.Template)
	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage) //Inline the Css to make more readable by other clients...
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.plain.tmpl", m.Templates, msg.Template)
	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

func (m *Mail) inlineCSS(s string) (string, error) { // Using the premailer pkg in this func...
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
