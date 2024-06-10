
# Emailer

Small email service written in GO


## Roadmap

* [ ]  Add testing
* [x]  Support attachments
* [x]  Support multiple target emails
* [ ]  Support Gmail GCP API
* [ ]  Support Microsoft Graph API
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

```json
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
        "from": "sender@gmail.com",
        "password": "p4$$w0rd"
      }
}
```

## Deployment

To deploy this project run

```go
  go build
  ./emailer
```

