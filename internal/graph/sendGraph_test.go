package graph

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kermeth/emailer/internal/send"
	"github.com/microsoftgraph/msgraph-sdk-go/models"
)

func TestDecodeGraphRequest_Success(t *testing.T) {
	payload := `{
		"to": ["to@example.com"],
		"cc": ["cc@example.com"],
		"bcc": ["bcc@example.com"],
		"subject": "Hello",
		"body": "<b>World</b>",
		"attachments": [{"name":"a.txt","data":"ABC"}],
		"configuration": {"tenantId":"t","appId":"a","secret":"s","from":"sender@example.com"}
	}`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(payload))
	req, err := decodeGraphRequest(r)
	if err != nil {
		t.Fatalf("decodeGraphRequest() error = %v", err)
	}
	if got, want := len(req.To), 1; got != want {
		t.Fatalf("To len = %d, want %d", got, want)
	}
	if req.Subject != "Hello" || req.Body != "<b>World</b>" {
		t.Fatalf("unexpected subject/body: %q / %q", req.Subject, req.Body)
	}
	if got, want := len(req.Attachments), 1; got != want {
		t.Fatalf("Attachments len = %d, want %d", got, want)
	}
	if req.Configuration.From != "sender@example.com" {
		t.Fatalf("unexpected From: %s", req.Configuration.From)
	}
}

func TestDecodeGraphRequest_InvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{"))
	_, err := decodeGraphRequest(r)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestBuildGraphRecipients_FiltersEmptyAndSetsAddress(t *testing.T) {
	in := []string{"foo@example.com", "", "bar@example.com"}
	recips := buildGraphRecipients(in)
	if got, want := len(recips), 2; got != want {
		t.Fatalf("recipients len = %d, want %d", got, want)
	}
	// verify addresses
	for i, rcp := range recips {
		em := rcp.GetEmailAddress()
		if em == nil || em.GetAddress() == nil || *em.GetAddress() == "" {
			t.Fatalf("recipient %d has no address", i)
		}
	}
}

func TestBuildGraphAttachments_SetsNameAndBytes(t *testing.T) {
	atts := []send.Attachment{{Name: "file.bin", Data: "XYZ"}}
	built := buildGraphAttachments(atts)
	if got, want := len(built), 1; got != want {
		t.Fatalf("attachments len = %d, want %d", got, want)
	}
	// Assert concrete type exposes expected getters
	fa, ok := built[0].(models.FileAttachmentable)
	if !ok {
		t.Fatalf("attachment is not a FileAttachmentable: %T", built[0])
	}
	if fa.GetName() == nil || *fa.GetName() != "file.bin" {
		t.Fatalf("unexpected attachment name: %v", fa.GetName())
	}
	if string(fa.GetContentBytes()) != "XYZ" { // data is used as-is
		t.Fatalf("unexpected content bytes: %v", fa.GetContentBytes())
	}
}

func TestHandler_BadRequest_OnInvalidJSON(t *testing.T) {
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("{"))
	w := httptest.NewRecorder()
	Handler(w, r)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandler_InternalServerError_OnSendFailure(t *testing.T) {
	// Valid JSON but missing/invalid credentials so sendEmail should fail quickly
	reqBody := Request{
		To:      []string{"to@example.com"},
		Subject: "S",
		Body:    "B",
		Configuration: Config{
			TenantId: "", // invalid
			AppId:    "",
			Secret:   "",
			From:     "sender@example.com",
		},
	}
	b, _ := json.Marshal(reqBody)
	r := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	w := httptest.NewRecorder()
	Handler(w, r)
	if w.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
	}
}
