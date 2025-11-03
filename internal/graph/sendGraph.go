package graph

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"

	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	azidentity "github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/kermeth/emailer/internal/send"
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
	Sender   string `json:"sender"`
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

	// Get access token for Microsoft Graph
	token, err := cred.GetToken(context.Background(), policy.TokenRequestOptions{
		Scopes: []string{"https://graph.microsoft.com/.default"},
	})
	if err != nil {
		return fmt.Errorf("failed to get token: %w", err)
	}

	// Build the email message
	message := map[string]interface{}{
		"message": map[string]interface{}{
			"subject": request.Subject,
			"body": map[string]string{
				"contentType": "HTML",
				"content":     request.Body,
			},
			"toRecipients": buildRecipients(request.To),
		},
		"saveToSentItems": true,
	}

	// Add CC recipients if present
	if len(request.Cc) > 0 {
		message["message"].(map[string]interface{})["ccRecipients"] = buildRecipients(request.Cc)
	}

	// Add BCC recipients if present
	if len(request.Bcc) > 0 {
		message["message"].(map[string]interface{})["bccRecipients"] = buildRecipients(request.Bcc)
	}

	// Add attachments if present
	if len(request.Attachments) > 0 {
		message["message"].(map[string]interface{})["attachments"] = buildAttachments(request.Attachments)
	}

	// Convert to JSON
	jsonData, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Log the request for debugging
	slog.Debug("Sending email", "sender", request.Configuration.Sender, "url", fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", request.Configuration.Sender))

	// Send the email via Graph API
	userEmail := request.Configuration.Sender
	apiURL := fmt.Sprintf("https://graph.microsoft.com/v1.0/users/%s/sendMail", userEmail)
	req, err := http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token.Token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to send email: status %d, body: %s", resp.StatusCode, string(body))
	}

	return nil
}

func buildRecipients(emails []string) []map[string]interface{} {
	recipients := make([]map[string]interface{}, 0, len(emails))
	for _, email := range emails {
		if email != "" {
			recipients = append(recipients, map[string]interface{}{
				"emailAddress": map[string]string{
					"address": email,
				},
			})
		}
	}
	return recipients
}

func buildAttachments(attachments []send.Attachment) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(attachments))
	for _, att := range attachments {
		result = append(result, map[string]interface{}{
			"@odata.type":  "#microsoft.graph.fileAttachment",
			"name":         att.Name,
			"contentBytes": att.Data, // Should be base64 encoded
		})
	}
	return result
}
