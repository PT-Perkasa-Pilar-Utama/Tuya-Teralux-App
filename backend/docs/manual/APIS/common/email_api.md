# Email API

## Send Email

This endpoint allows sending emails via the configured SMTP server using server-side HTML templates.

### Validations
- `to`: Must be a valid email address (array of strings).
- `subject`: Must not be empty.
- `template`: Must be a valid template name (e.g., "summary", "test").

### Request

- **URL** : `/api/email/send`
- **Method** : `POST`
- **Auth required** : Yes (Bearer Token)
- **Permissions** : Admin or authorized user

#### Header

```http
Authorization: Bearer <token>
Content-Type: application/json
```

#### Body

```json
{
    "to": ["recipient@example.com"],
    "subject": "Meeting Summary",
    "template": "summary"
}
```

### Response

#### Success (200 OK)

```json
{
    "status": true,
    "message": "Email sent successfully"
}
```

#### Error (400 Bad Request)

```json
{
    "status": false,
    "message": "Validation Error",
    "details": [
        { "field": "to", "message": "recipient list cannot be empty" }
    ]
}
```

#### Error (500 Internal Server Error)

```json
{
    "status": false,
    "message": "Failed to send email",
    "details": "SMTP server unavailable: connection refused"
}
```
