package graph

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	"context"

	azidentity "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/kermeth/emailer/internal/send"
	abstractions "github.com/microsoft/kiota-abstractions-go"
	msgraphsdk "github.com/microsoftgraph/msgraph-sdk-go"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
	"github.com/microsoftgraph/msgraph-sdk-go/users"
)

type Request struct {
	To            []string          `json:"to"`
	Cc            []string          `json:"cc"`
	Bcc           []string          `json:"bcc"`
	Subject       string            `json:"subject"`
	Body          string            `json:"body"`
	Attachments   []send.Attachment `json:"attachments"`
	Configuration Config            `json:"configuration"`
}

type Config struct {
	TenantId string `json:"tenantId"`
	AppId    string `json:"appId"`
	Secret   string `json:"secret"`
	From     string `json:"from"`
}

func Handler(writer http.ResponseWriter, request *http.Request) {
	graphRequest, err := decodeGraphRequest(request)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		slog.Error("failed to decode email request", "Error", err)
		return
	}
	err = graphRequest.sendEmail()
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		slog.Error("failed to send email request", "Error", err)
		return
	}
	slog.Info("Email sent")
	writer.WriteHeader(http.StatusOK)
}

func decodeGraphRequest(request *http.Request) (*Request, error) {
	var graphRequest Request
	err := json.NewDecoder(request.Body).Decode(&graphRequest)
	if err != nil {
		return nil, err
	}
	return &graphRequest, nil
}

func (request *Request) sendEmail() error {
	// Create credential using client secret
	cred, err := azidentity.NewClientSecretCredential(
		request.Configuration.TenantId,
		request.Configuration.AppId,
		request.Configuration.Secret,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to create credential: %w", err)
	}

	// Create Graph client
	client, err := msgraphsdk.NewGraphServiceClientWithCredentials(cred, []string{
		"https://graph.microsoft.com/.default",
	})
	if err != nil {
		return fmt.Errorf("failed to create graph client: %w", err)
	}

	// Build message
	message := models.NewMessage()
	subject := request.Subject
	message.SetSubject(&subject)

	// Set body
	body := models.NewItemBody()
	bodyContent := request.Body
	body.SetContent(&bodyContent)
	contentType := models.HTML_BODYTYPE
	body.SetContentType(&contentType)
	message.SetBody(body)

	// Set recipients
	message.SetToRecipients(buildGraphRecipients(request.To))

	if len(request.Cc) > 0 {
		message.SetCcRecipients(buildGraphRecipients(request.Cc))
	}

	if len(request.Bcc) > 0 {
		message.SetBccRecipients(buildGraphRecipients(request.Bcc))
	}

	// Add attachments if present
	if len(request.Attachments) > 0 {
		message.SetAttachments(buildGraphAttachments(request.Attachments))
	}

	// Create send mail request body
	sendMailBody := users.NewItemSendMailPostRequestBody()
	sendMailBody.SetMessage(message)
	saveToSentItems := true
	sendMailBody.SetSaveToSentItems(&saveToSentItems)

	// Add request options with logging
	requestConfig := &users.ItemSendMailRequestBuilderPostRequestConfiguration{
		Options: []abstractions.RequestOption{},
	}

	slog.Info("Attempting to send email",
		"sender", request.Configuration.From,
		"recipients", request.To,
		"subject", request.Subject)

	// Send the email
	err = client.Users().
		ByUserId(request.Configuration.From).
		SendMail().
		Post(context.Background(), sendMailBody, requestConfig)

	if err != nil {
		// Try to extract more error details
		slog.Error("Graph API error details",
			"error", err.Error(),
			"sender", request.Configuration.From,
			"type", fmt.Sprintf("%T", err))
		return fmt.Errorf("failed to send email: %w", err)
	}

	slog.Info("Email sent successfully via Graph API")
	return nil
}

func buildGraphRecipients(emails []string) []models.Recipientable {
	recipients := make([]models.Recipientable, 0, len(emails))
	for _, email := range emails {
		if email != "" {
			recipient := models.NewRecipient()
			emailAddress := models.NewEmailAddress()
			emailAddr := email
			emailAddress.SetAddress(&emailAddr)
			recipient.SetEmailAddress(emailAddress)
			recipients = append(recipients, recipient)
		}
	}
	return recipients
}

func buildGraphAttachments(attachments []send.Attachment) []models.Attachmentable {
	result := make([]models.Attachmentable, 0, len(attachments))
	for _, att := range attachments {
		attachment := models.NewFileAttachment()
		name := att.Name
		attachment.SetName(&name)

		// The Data field should already be base64 encoded
		contentBytes := []byte(att.Data)
		attachment.SetContentBytes(contentBytes)

		result = append(result, attachment)
	}
	return result
}
