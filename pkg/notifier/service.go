package notifier

import "go.uber.org/zap"

// Notifier an interface for sending notifications
type Notifier interface {
	Info(header, body string, data map[string]interface{})
	Error(header, body string, err error, data map[string]interface{})
}

type Options struct {
	Webhook *WebhookOptions
}

func New(options Options, logger *zap.SugaredLogger) Notifier {
	if options.Webhook == nil || len(options.Webhook.Url) == 0 {
		return Disabled{}
	} else {
		return NewWebhook(*options.Webhook, logger)
	}
}
