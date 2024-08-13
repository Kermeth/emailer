package gmail

import (
	"context"
	"google.golang.org/api/gmail/v1"
	"google.golang.org/api/option"
	"log"
	"net/http"
)

func Handler(writer http.ResponseWriter, request *http.Request) {
	ctx := context.Background()
	srv, err := gmail.NewService(ctx, option.WithCredentialsJSON([]byte("ignore/credentials.json")))
	if err != nil {
		log.Fatalf("Unable to retrieve Gmail client: %v", err)
	}

	// TODO - Add the email message to the request body
	headerTo := &gmail.MessagePartHeader{Name: "To", Value: "email@gmail.com"}
	headerSubject := &gmail.MessagePartHeader{Name: "Subject", Value: "Hello World"}
	srv.Users.Messages.Send("me", &gmail.Message{
		Payload: &gmail.MessagePart{
			Body: &gmail.MessagePartBody{
				Data: "Hello World",
			},
			Headers: []*gmail.MessagePartHeader{headerTo, headerSubject},
		},
	})
	writer.WriteHeader(http.StatusOK)
}
