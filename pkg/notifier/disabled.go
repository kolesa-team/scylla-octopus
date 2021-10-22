package notifier

// Disabled is a notifier that doesn't send any notifications
type Disabled struct {
}

func (d Disabled) Info(header, body string, data map[string]interface{}) {
}

func (d Disabled) Error(header, body string, err error, data map[string]interface{}) {
}
