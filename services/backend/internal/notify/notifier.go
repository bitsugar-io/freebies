package notify

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/pinpoint"
	"github.com/aws/aws-sdk-go-v2/service/pinpoint/types"
)

// NewConsole creates a new console notifier
func NewConsole(
	logger *slog.Logger,
) *ConsoleNotifier {
	return &ConsoleNotifier{
		logger: logger,
	}
}

// SendSMS logs the SMS to console
func (c *ConsoleNotifier) SendSMS(
	_ context.Context,
	toE164 string,
	body string,
) (string, error) {
	c.logger.Info("NOTIFY_SENT", "phone", toE164, "body", body)
	return "console-msg-id", nil
}

// NewPinpoint creates a new Pinpoint notifier
func NewPinpoint(
	ctx context.Context,
	logger *slog.Logger,
	appID string,
	senderID string,
	endpoint string,
) (*PinpointNotifier, error) {
	var cfg aws.Config
	var err error
	cfg, err = config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Override endpoint if provided (for LocalStack, though Pinpoint isn't well-supported)
	if endpoint != "" {
		cfg.BaseEndpoint = aws.String(endpoint)
	}

	client := pinpoint.NewFromConfig(cfg)

	return &PinpointNotifier{
		client:   client,
		appID:    appID,
		senderID: senderID,
		logger:   logger,
	}, nil
}

// SendSMS sends an SMS via Pinpoint
func (p *PinpointNotifier) SendSMS(
	ctx context.Context,
	toE164 string,
	body string,
) (string, error) {
	input := &pinpoint.SendMessagesInput{
		ApplicationId: aws.String(p.appID),
		MessageRequest: &types.MessageRequest{
			Addresses: map[string]types.AddressConfiguration{
				toE164: {
					ChannelType: types.ChannelTypeSms,
				},
			},
			MessageConfiguration: &types.DirectMessageConfiguration{
				SMSMessage: &types.SMSMessage{
					Body:              aws.String(body),
					MessageType:       types.MessageTypeTransactional,
					OriginationNumber: aws.String(p.senderID),
				},
			},
		},
	}

	var output *pinpoint.SendMessagesOutput
	var err error
	output, err = p.client.SendMessages(ctx, input)
	if err != nil {
		return "", fmt.Errorf("failed to send SMS: %w", err)
	}

	// Extract message ID from response
	var msgID string
	if output.MessageResponse != nil && output.MessageResponse.Result != nil {
		for _, result := range output.MessageResponse.Result {
			if result.MessageId != nil {
				msgID = *result.MessageId
				break
			}
		}
	}

	return msgID, nil
}

// New creates the appropriate notifier based on environment
func New(
	ctx context.Context,
	logger *slog.Logger,
	environment string,
	appID string,
	senderID string,
) Notifier {
	if environment == "local" || appID == "" {
		return NewConsole(logger)
	}

	var notifier *PinpointNotifier
	var err error
	notifier, err = NewPinpoint(ctx, logger, appID, senderID, "")
	if err != nil {
		logger.Warn("Failed to create Pinpoint notifier, falling back to console", "error", err)
		return NewConsole(logger)
	}

	return notifier
}
