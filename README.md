
# Emailer

Small email service written in GO

## Run on Docker
Image is already published on Docker Hub [here](https://hub.docker.com/r/kermeth/emailer)

```bash 
docker run kermeth/emailer
```

## Roadmap

* [ ]  Add testing
* [x]  Support attachments
* [x]  Support multiple target emails
* [ ]  Support Gmail GCP API
* [x]  Support Microsoft Graph API
* [ ]  Support Amazon SES
* [ ]  Retry Mechanism
* [ ]  Rate Limiting
* [ ]  Read message from RabbitMQ


## API Reference

#### Healthcheck

```http
  GET /health
```

#### Send email

```http
  POST /smtp/send
```

```json lines
{
      "to": [
        "receiver@gmail.com"
      ],
      "cc": [],
      "bcc": [],
      "subject": "Test email",
      "body": "This is <b>a</b> test email",
      "attachments": [
        {
          "name": "test.txt",
          "data": "VGhpcyBpcyBhIHRlc3QgYXR0YWNobWVudA=="
        }
      ],
      "configuration": {
        "host": "smtp.gmail.com",
        "port": 587,
        "alias": "John Doe", // Optional
        "loginType": "plain|login", // Optional (default: plain)
        "from": "sender@gmail.com",
        "password": "p4$$w0rd"
      }
}
```

#### Send email via Microsoft Graph

```http
  POST /graph/send
```

```json lines
{
      "to": [
        "receiver@contoso.com"
      ],
      "cc": [],
      "bcc": [],
      "subject": "Test email (Graph)",
      "body": "This is <b>a</b> test email sent via Microsoft Graph",
      "attachments": [
        {
          "name": "test.txt",
          "data": "VGhpcyBpcyBhIHRlc3QgYXR0YWNobWVudA=="
        }
      ],
      "configuration": {
        "tenantId": "00000000-0000-0000-0000-000000000000",
        "appId": "11111111-1111-1111-1111-111111111111",
        "secret": "your-client-secret",
        "from": "sender@contoso.com" // UPN (email) or user id to send as
      }
}
```

Notes:
- Body is sent as HTML.
- The Azure AD app must have Microsoft Graph Application permission `Mail.Send` and admin consent granted.
- The app must be allowed to send as the specified `from` user/mailbox in your tenant.
- Attachments `data` should be the Base64 content of the file.

## Deployment

To deploy this project run

```go
  go build
  ./emailer
```

## Changelog

- 0.1.1
  - Allow alias for sender
- 0.1.0
  - Initial Release
  - SMTP Support
  - Attachments Support

