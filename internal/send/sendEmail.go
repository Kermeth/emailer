package send

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
)

type EmailRequest struct {
	To            []string          `json:"to"`
	Cc            []string          `json:"cc"`
	Bcc           []string          `json:"bcc"`
	Subject       string            `json:"subject"`
	Body          string            `json:"body"`
	Attachments   []Attachment      `json:"attachments"`
	Configuration SMTPConfiguration `json:"configuration"`
}

type Attachment struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type SMTPConfiguration struct {
	Host      string    `json:"host"`
	Port      int       `json:"port"`
	Alias     string    `json:"alias,omitempty"`
	LoginType LoginType `json:"loginType,omitempty"`
	From      string    `json:"from"`
	Password  string    `json:"password"`
}

type LoginType string

const (
	Plain LoginType = "plain"
	Login LoginType = "login"
)

func Handler(writer http.ResponseWriter, request *http.Request) {
	emailRequest, err := decodeEmailRequest(request)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		slog.Error("failed to decode email request", "Error", err)
		return
	}
	err = emailRequest.sendEmail()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to send email", "Error:", err)
		return
	}
	slog.Info("Email sent")
	writer.WriteHeader(http.StatusOK)
}

func decodeEmailRequest(request *http.Request) (*EmailRequest, error) {
	var emailRequest EmailRequest
	err := json.NewDecoder(request.Body).Decode(&emailRequest)
	if err != nil {
		return nil, err
	}
	return &emailRequest, nil
}

func (request *EmailRequest) sendEmail() error {
	auth := smtp.PlainAuth(
		"",
		request.Configuration.From,
		request.Configuration.Password,
		request.Configuration.Host)
	if request.Configuration.LoginType == Login {
		slog.Info("Using login auth")
		auth = LoginAuth(request.Configuration.From, request.Configuration.Password)
	}
	server := request.Configuration.Host + ":" + strconv.Itoa(request.Configuration.Port)
	err := smtp.SendMail(server, auth, request.Configuration.From, request.To, request.toBytes())
	if err != nil {
		return err
	}
	return nil
}

func (request *EmailRequest) toBytes() []byte {
	buffer := bytes.NewBuffer(nil)
	withAttachments := len(request.Attachments) > 0
	from := mail.Address{
		Name:    request.Configuration.Alias,
		Address: request.Configuration.From,
	}
	buffer.WriteString(fmt.Sprintf("From: %s\r\n", from.String()))
	buffer.WriteString(fmt.Sprintf("Subject: %s\r\n", request.Subject))
	buffer.WriteString(fmt.Sprintf("To: %s\r\n", strings.Join(request.To, ",")))
	if len(request.Cc) > 0 {
		buffer.WriteString(fmt.Sprintf("Cc: %s\r\n", strings.Join(request.Cc, ",")))
	}
	if len(request.Bcc) > 0 {
		buffer.WriteString(fmt.Sprintf("Bcc: %s\r\n", strings.Join(request.Bcc, ",")))
	}
	buffer.WriteString("MIME-Version: 1.0\r\n")
	writer := multipart.NewWriter(buffer)
	boundary := writer.Boundary()
	if withAttachments {
		buffer.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n", boundary))
		buffer.WriteString(fmt.Sprintf("--%s\r\n", boundary))
	}
	// Write HTML body
	buffer.WriteString("Content-Type: text/html; charset=utf-8\r\n")
	buffer.WriteString("\r\n") // Empty line to separate headers from body RFC 822-style email
	buffer.WriteString(request.Body)
	// Write attachments
	if withAttachments {
		for _, attachment := range request.Attachments {
			v := []byte(attachment.Data)
			buffer.WriteString(fmt.Sprintf("\n\n--%s\n", boundary))
			buffer.WriteString(fmt.Sprintf("Content-Type: %s\n", http.DetectContentType(v)))
			buffer.WriteString("Content-Transfer-Encoding: base64\n")
			buffer.WriteString(fmt.Sprintf("Content-Disposition: attachment; filename=%s\n", attachment.Name))

			b := make([]byte, base64.StdEncoding.EncodedLen(len(v)))
			base64.StdEncoding.Encode(b, v)
			buffer.Write(b)
			buffer.WriteString(fmt.Sprintf("\n--%s", boundary))
		}
		buffer.WriteString("--")
	}
	return buffer.Bytes()
}
