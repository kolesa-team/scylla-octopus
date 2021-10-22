package notifier

import (
	"errors"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestWebhook_getMessage_NoError(t *testing.T) {
	w := &Webhook{options: WebhookOptions{Url: "test-url"}}
	actualMessage := w.getMessage("Test header", "This is a test notification message", nil, map[string]interface{}{
		"extraField": "extraValue",
	})
	expectedMessage := `🐙 <b>scylla-octopus</b>

<b>Test header</b>

This is a test notification message

<pre>{
  "extraField": "extraValue"
}</pre>`

	require.Equal(t, expectedMessage, actualMessage)
}

func TestWebhook_getMessage_WithError(t *testing.T) {
	w := &Webhook{options: WebhookOptions{Url: "test-url"}}
	actualMessage := w.getMessage(
		"Test header",
		"",
		errors.New("this is a test error"),
		map[string]interface{}{
			"extraField": "extraValue",
		},
	)
	expectedMessage :=
		`🐙 <b>scylla-octopus</b> 🔥

<b>Test header</b>

Error:
<pre>this is a test error</pre>

<pre>{
  "extraField": "extraValue"
}</pre>`

	require.Equal(t, expectedMessage, actualMessage)
}
