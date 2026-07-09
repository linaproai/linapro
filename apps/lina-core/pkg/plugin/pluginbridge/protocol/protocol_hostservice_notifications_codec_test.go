// This file tests notifications host service request and response codec round
// trips.

package protocol

import "testing"

// TestHostServiceNotificationsSendRoundTrip verifies notifications send requests and
// responses preserve message metadata, recipients, and payload JSON.
func TestHostServiceNotificationsSendRoundTrip(t *testing.T) {
	original := &HostServiceNotificationsSendRequest{
		Title:            "同步完成",
		Content:          "订单同步已完成",
		SourceType:       "plugin",
		SourceID:         "job-1",
		CategoryCode:     "other",
		RecipientUserIDs: []int64{1, 2},
		PayloadJSON:      []byte(`{"scope":"orders"}`),
	}

	requestData := MarshalHostServiceNotificationsSendRequest(original)
	request, err := UnmarshalHostServiceNotificationsSendRequest(requestData)
	if err != nil {
		t.Fatalf("request unmarshal failed: %v", err)
	}
	if request.Title != original.Title || request.Content != original.Content {
		t.Fatalf("request: got %#v want %#v", request, original)
	}
	if len(request.RecipientUserIDs) != 2 || request.RecipientUserIDs[1] != 2 {
		t.Fatalf("recipientUserIDs: got %#v", request.RecipientUserIDs)
	}
	if string(request.PayloadJSON) != string(original.PayloadJSON) {
		t.Fatalf("payloadJson: got %s want %s", string(request.PayloadJSON), string(original.PayloadJSON))
	}

	responseOriginal := &HostServiceNotificationsSendResponse{
		MessageID:     101,
		DeliveryCount: 2,
	}
	responseData := MarshalHostServiceNotificationsSendResponse(responseOriginal)
	response, err := UnmarshalHostServiceNotificationsSendResponse(responseData)
	if err != nil {
		t.Fatalf("response unmarshal failed: %v", err)
	}
	if response.MessageID != responseOriginal.MessageID || response.DeliveryCount != responseOriginal.DeliveryCount {
		t.Fatalf("response: got %#v want %#v", response, responseOriginal)
	}
}
