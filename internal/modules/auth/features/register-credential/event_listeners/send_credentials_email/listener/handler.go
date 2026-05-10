// Package listener contains the event listener responsible for emailing
// the set-password link to a user whose credential was just created in
// pending state.
//
// Flow (not yet implemented):
//
//  1. Consumes the `auth.credential.pending` event published by
//     create_specialist_credential after inserting the credential row and
//     generating the set-password token.
//  2. Calls an email delivery service (SES, SendGrid, or equivalent), passing:
//     - recipient email (from payload)
//     - set-password deep link containing the token (from payload)
//     - localized template identifier based on role
//  3. Records an audit log entry on both success and failure.
//  4. Failures surface as handler errors so the SQS consumer keeps the
//     message invisible until the retry window expires and eventually
//     routes it to the DLQ.
//
// The placeholder lives at:
//
//	internal/modules/auth/features/register-credential/event_listeners/send_credentials_email/
//	├── listener/
//	│   ├── handler.go            <- this file
//	│   ├── new_handler.go
//	│   ├── interface.go          <- EmailDelivery contract
//	│   ├── dto.go                <- mirror of AuthCredentialPending payload
//	│   ├── constants.go
//	│   └── mocks/
//	└── adapters/
//	    ├── inbound/sqs/manager.go
//	    └── outbound/email/       <- SES or SendGrid gateway
package listener
