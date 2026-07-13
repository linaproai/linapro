// This file defines the optional email delivery bridge used by the host notify
// domain when channel_type=email. The concrete implementation is provided by
// linapro-mail-core (or tests); the host must not depend on SMTP plugins.

package notifycap

import (
	"context"
	"sync"

	"github.com/gogf/gf/v2/errors/gerror"
)

// EmailDelivery delivers one email notification payload through the mail owner.
type EmailDelivery interface {
	// Deliver sends one email notification for the given recipients and content.
	// AccountID zero means the mail owner should resolve its default account.
	Deliver(ctx context.Context, in EmailDeliveryInput) (*EmailDeliveryResult, error)
}

// EmailDeliveryInput is the host-facing email delivery request.
type EmailDeliveryInput struct {
	// AccountID selects a mail-core Account; zero means default.
	AccountID int64
	// To contains recipient email addresses.
	To []string
	// Subject is the email subject.
	Subject string
	// Content is the email body (text/html is owner-defined).
	Content string
}

// EmailDeliveryResult describes one delivery outcome.
type EmailDeliveryResult struct {
	// ProviderMessageID is an optional transport message identifier.
	ProviderMessageID string
}

var emailDeliveryRegistry = struct {
	sync.RWMutex
	delivery EmailDelivery
}{}

// ProvideEmailDelivery registers the process-local email delivery implementation.
// Only one provider may be registered; subsequent calls return an error.
func ProvideEmailDelivery(delivery EmailDelivery) error {
	if delivery == nil {
		return gerror.New("notifycap: email delivery is nil")
	}
	emailDeliveryRegistry.Lock()
	defer emailDeliveryRegistry.Unlock()
	if emailDeliveryRegistry.delivery != nil {
		return gerror.New("notifycap: email delivery already registered")
	}
	emailDeliveryRegistry.delivery = delivery
	return nil
}

// EmailDeliveryOrNil returns the registered email delivery implementation, if any.
func EmailDeliveryOrNil() EmailDelivery {
	emailDeliveryRegistry.RLock()
	defer emailDeliveryRegistry.RUnlock()
	return emailDeliveryRegistry.delivery
}

// ResetEmailDeliveryForTest clears the registered email delivery. Tests only.
func ResetEmailDeliveryForTest() {
	emailDeliveryRegistry.Lock()
	defer emailDeliveryRegistry.Unlock()
	emailDeliveryRegistry.delivery = nil
}
