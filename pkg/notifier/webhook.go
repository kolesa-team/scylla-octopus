package notifier

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"html"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Webhook is an implementation of Notifier that sends HTML messages to a given URL
type Webhook struct {
	options    WebhookOptions
	httpClient *http.Client
	logger     *zap.SugaredLogger
}

type WebhookOptions struct {
	Url          string
	MessageField string
}

// Info sends an informational message
func (w *Webhook) Info(header, body string, data map[string]interface{}) {
	msg := w.getMessage(header, body, nil, data)
	err := w.sendMessage(msg)
	if err != nil {
		w.logger.Warnw("could not send message to webhook", "message", msg, "error", err)
	}
}

// Error sends an error message
func (w *Webhook) Error(header, body string, err error, data map[string]interface{}) {
	msg := w.getMessage(header, body, err, data)
	sendErr := w.sendMessage(msg)
	if sendErr != nil {
		w.logger.Warnw("could not send message to webhook", "message", msg, "error", sendErr)
	}
}

// Creates a notification message
func (w *Webhook) getMessage(header, body string, err error, extra map[string]interface{}) string {
	lines := []string{}
	prefix := "🐙 <b>scylla-octopus</b>"

	if err != nil {
		prefix = prefix + " 🔥" // an error needs some fire

		if len(body) == 0 {
			body = fmt.Sprintf("Error:\n<pre>%s</pre>", html.EscapeString(err.Error()))
		}
	}

	lines = append(lines, prefix, "")

	if len(header) > 0 {
		lines = append(
			lines,
			fmt.Sprintf("<b>%s</b>", header),
			"",
		)
	}

	lines = append(lines, body)

	if len(extra) > 0 {
		extraJsonBytes, _ := json.MarshalIndent(extra, "", "  ")
		extraJson := fmt.Sprintf("<pre>%s</pre>", string(extraJsonBytes))
		lines = append(lines, "", extraJson)
	}

	return strings.Join(lines, "\n")
}

func (w *Webhook) sendMessage(message string) error {
	resp, err := w.httpClient.PostForm(w.options.Url, url.Values{
		w.options.MessageField: []string{message},
	})
	if err != nil {
		return errors.Wrap(err, "could not send message to webhook")
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code from webhook: %d", resp.StatusCode)
	}

	return nil
}

func NewWebhook(options WebhookOptions, logger *zap.SugaredLogger) *Webhook {
	if len(options.MessageField) == 0 {
		options.MessageField = "message"
	}

	return &Webhook{
		logger:  logger.Named("webhook-notifier"),
		options: options,
		httpClient: &http.Client{
			Timeout: time.Second,
		},
	}
}
