package notify

import (
	"context"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/service/pinpoint"
)

// Notifier sends SMS notifications
type Notifier interface {
	SendSMS(
		ctx context.Context,
		toE164 string,
		body string,
	) (providerMsgID string, err error)
}

// ConsoleNotifier logs SMS to console (for local dev)
type ConsoleNotifier struct {
	logger *slog.Logger
}

// PinpointNotifier sends SMS via AWS Pinpoint
type PinpointNotifier struct {
	client   *pinpoint.Client
	appID    string
	senderID string
	logger   *slog.Logger
}
