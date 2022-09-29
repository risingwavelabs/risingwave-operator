package utils

// TODO: move WebhookType out of metrics package

// Defines an immutable type for a webhook. Use NewWebhookType to instantiate this.
type WebhookType struct {
	isValidating_ bool
}

// NewWebhookTypes returns an immutable webhookType.
func NewWebhookTypes(isValidating bool) WebhookType {
	return WebhookType{isValidating_: isValidating}
}

// IsValidating is true if wt is a validating webhook, else false.
func (wt *WebhookType) IsValidating() bool {
	return wt.isValidating_
}

// IsMutating is true if wt is a mutating webhook, else false.
func (wt *WebhookType) IsMutating() bool {
	return !wt.IsValidating()
}

func (wt *WebhookType) String() string {
	if wt.IsValidating() {
		return "validating"
	}
	return "mutating"
}

// Equal is true if wt is equivalent to other.
func (wt *WebhookType) Equal(other *WebhookType) bool {
	return wt.IsValidating() == other.IsValidating()
}
