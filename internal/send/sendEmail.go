package send

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/smtp"
	"strconv"
	"strings"
)

type EmailRequest struct {
	Api           string          `json:"api"`
	To            []string        `json:"to"`
	Cc            []string        `json:"cc"`
	Bcc           []string        `json:"bcc"`
	Subject       string          `json:"subject"`
	Body          string          `json:"body"`
	Attachments   []Attachment    `json:"attachments"`
	Configuration json.RawMessage `json:"configuration"`
}

type Attachment struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

type SMTPConfiguration struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	From     string `json:"from"`
	Password string `json:"password"`
}

func Handler(writer http.ResponseWriter, request *http.Request) {
	emailRequest, err := decodeEmailRequest(request)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		slog.Error("failed to decode email request", "Error", err)
		return
	}
	switch emailRequest.Api {
	case "smtp":
		config, err := decodeSMTPConfiguration(emailRequest.Configuration)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			slog.Error("failed to decode smtp configuration", "Error", err)
			return
		}
		err = config.sendEmail(emailRequest)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			slog.Error("failed to send email", "Error:", err)
			return
		}
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

func decodeSMTPConfiguration(configuration json.RawMessage) (*SMTPConfiguration, error) {
	var smtpConfig SMTPConfiguration
	err := json.Unmarshal(configuration, &smtpConfig)
	if err != nil {
		return nil, err
	}
	return &smtpConfig, nil
}

func (config *SMTPConfiguration) sendEmail(request *EmailRequest) error {
	auth := smtp.PlainAuth("", config.From, config.Password, config.Host)
	server := config.Host + ":" + strconv.Itoa(config.Port)
	err := smtp.SendMail(server, auth, config.From, request.To, request.toBytes())
	if err != nil {
		return err
	}
	return nil
}

func (e *EmailRequest) toBytes() []byte {
	buffer := bytes.NewBuffer(nil)
	withAttachments := len(e.Attachments) > 0
	buffer.WriteString(fmt.Sprintf("Subject: %s\n", e.Subject))
	buffer.WriteString(fmt.Sprintf("To: %s\n", strings.Join(e.To, ",")))
	if len(e.Cc) > 0 {
		buffer.WriteString(fmt.Sprintf("Cc: %s\n", strings.Join(e.Cc, ",")))
	}
	if len(e.Bcc) > 0 {
		buffer.WriteString(fmt.Sprintf("Bcc: %s\n", strings.Join(e.Bcc, ",")))
	}
	buffer.WriteString("MIME-Version: 1.0\n")
	writer := multipart.NewWriter(buffer)
	boundary := writer.Boundary()
	if withAttachments {
		buffer.WriteString(fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\n", boundary))
		buffer.WriteString(fmt.Sprintf("--%s\n", boundary))
	}
	// Write HTML body
	buffer.WriteString("Content-Type: text/html; charset=utf-8\n")
	buffer.WriteString(e.Body)
	// Write attachments
	if withAttachments {
		for _, attachment := range e.Attachments {
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
