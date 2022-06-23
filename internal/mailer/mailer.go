package mailer

import (
	"bytes"
	"embed"
	"html/template"
	"time"

	mail "github.com/xhit/go-simple-mail/v2"
)

// Below we declare a new variable with the type embed.FS (embedded file system) to hold
// our email templates. This has a comment directive in the format `//go:embed <path>`
// IMMEDIATELY ABOVE it, which indicates to Go that we want to store the contents of the
// ./templates directory in the templateFS embedded file system variable.

//go:embed "templates"
var templateFS embed.FS

// Define a Mailer struct which contains a mail.Dialer instance (used to connect to a
// SMTP server) and the sender information for your emails (the name and address you
// want the email to be from, such as "Alice Smith <alice@example.com>").
type Mailer struct {
	server *mail.SMTPServer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	server := mail.NewSMTPClient()
	server.Host = host
	server.Port = port
	server.Username = username
	server.Password = password
	server.KeepAlive = false
	server.ConnectTimeout = 5 * time.Second
	server.SendTimeout = 5 * time.Second

	// Return a Mailer instance containing the dialer and sender information.
	return Mailer{
		server: server,
		sender: sender,
	}
}

// Define a Send() method on the Mailer type. This takes the recipient email address
// as the first parameter, the name of the file containing the templates and any
// dynamic data for the templates as an interface{} parameter.
func (m Mailer) Send(recipient, templateFile string, data interface{}) error {
	// Use the ParseFS() method to parse the required template file from the embedded file system
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}

	// Execute the named template "subject", passing in the dynamic data and storing the
	// result in a bytes.Buffer variable.
	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return err
	}

	// Follow the same pattern to execute the "plainBody" template and store the result
	// in the plainBody variable.
	plainBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(plainBody, "plainBody", data)
	if err != nil {
		return err
	}

	// And likewise with the "htmlBody" template.
	htmlBody := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data)
	if err != nil {
		return err
	}

	// Setup email message
	email := mail.NewMSG()
	email.SetFrom(m.sender)
	email.AddTo(recipient)
	email.SetSubject(subject.String())
	email.SetBody(mail.TextPlain, plainBody.String())
	email.AddAlternative(mail.TextHTML, htmlBody.String())

	// Connect to email server
	client, err := m.server.Connect()
	if err != nil {
		return err
	}

	// Try sending the email up to three times before aborting and returning the final
	// error. We sleep for 500 milliseconds between each attempt.
	for i := 1; i <= 3; i++ {
		err = email.Send(client)

		// If everything worked, return nil
		if nil == err {
			return nil
		}

		// If it didn't work, sleep for a short time and retry
		time.Sleep(500 * time.Millisecond)
	}

	return err
}
