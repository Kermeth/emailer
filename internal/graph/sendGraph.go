package graph

import (
	"encoding/json"
	"log/slog"
	"net/http"

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
	return nil
}
