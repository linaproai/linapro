// This file defines notification-service business error codes and their i18n
// metadata.

package notify

import (
	"github.com/gogf/gf/v2/errors/gcode"

	"lina-core/pkg/bizerr"
)

var (
	// CodeNotifyUserNotFound reports that the target inbox user does not exist.
	CodeNotifyUserNotFound = bizerr.MustDefine(
		"NOTIFY_USER_NOT_FOUND",
		"User does not exist",
		gcode.CodeNotFound,
	)
	// CodeNotifyChannelTypeUnsupported reports that the requested channel type is not supported.
	CodeNotifyChannelTypeUnsupported = bizerr.MustDefine(
		"NOTIFY_CHANNEL_TYPE_UNSUPPORTED",
		"Notification channel type {channelType} is not supported",
		gcode.CodeInvalidParameter,
	)
	// CodeNotifyTitleRequired reports that a notification title is required.
	CodeNotifyTitleRequired = bizerr.MustDefine(
		"NOTIFY_TITLE_REQUIRED",
		"Notification title cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeNotifyInboxRecipientRequired reports that an inbox notification has no recipient users.
	CodeNotifyInboxRecipientRequired = bizerr.MustDefine(
		"NOTIFY_INBOX_RECIPIENT_REQUIRED",
		"Inbox notification recipient users cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeNotifyChannelKeyRequired reports that a notification channel key is required.
	CodeNotifyChannelKeyRequired = bizerr.MustDefine(
		"NOTIFY_CHANNEL_KEY_REQUIRED",
		"Notification channel key cannot be empty",
		gcode.CodeInvalidParameter,
	)
	// CodeNotifyChannelUnavailable reports that the requested notification channel is missing or disabled.
	CodeNotifyChannelUnavailable = bizerr.MustDefine(
		"NOTIFY_CHANNEL_UNAVAILABLE",
		"Notification channel does not exist or is disabled",
		gcode.CodeNotFound,
	)
	// CodeNotifyChannelQueryFailed reports that a notification channel cannot be queried.
	CodeNotifyChannelQueryFailed = bizerr.MustDefine(
		"NOTIFY_CHANNEL_QUERY_FAILED",
		"Failed to query notification channel",
		gcode.CodeInternalError,
	)
	// CodeNotifyRecipientQueryFailed reports that inbox notification recipients cannot be queried.
	CodeNotifyRecipientQueryFailed = bizerr.MustDefine(
		"NOTIFY_RECIPIENT_QUERY_FAILED",
		"Failed to query inbox notification recipients",
		gcode.CodeInternalError,
	)
	// CodeNotifyMessageCreateFailed reports that the notification message row cannot be created.
	CodeNotifyMessageCreateFailed = bizerr.MustDefine(
		"NOTIFY_MESSAGE_CREATE_FAILED",
		"Failed to create notification message",
		gcode.CodeInternalError,
	)
	// CodeNotifyDeliveryCreateFailed reports that a notification delivery row cannot be created.
	CodeNotifyDeliveryCreateFailed = bizerr.MustDefine(
		"NOTIFY_DELIVERY_CREATE_FAILED",
		"Failed to create notification delivery",
		gcode.CodeInternalError,
	)
	// CodeNotifyPayloadMarshalFailed reports that notification extension payload cannot be serialized.
	CodeNotifyPayloadMarshalFailed = bizerr.MustDefine(
		"NOTIFY_PAYLOAD_MARSHAL_FAILED",
		"Failed to serialize notification payload",
		gcode.CodeInternalError,
	)
)
