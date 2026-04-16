// This file tests cache host service request and response codec round trips.

package pluginbridge

import "testing"

func TestHostServiceCacheGetResponseRoundTrip(t *testing.T) {
	original := &HostServiceCacheGetResponse{
		Found: true,
		Value: &HostServiceCacheValue{
			ValueKind: HostServiceCacheValueKindString,
			Value:     "alpha",
			IntValue:  0,
			ExpireAt:  "2026-04-15T12:00:00Z",
		},
	}

	data := MarshalHostServiceCacheGetResponse(original)
	decoded, err := UnmarshalHostServiceCacheGetResponse(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if !decoded.Found {
		t.Fatal("found: expected true")
	}
	if decoded.Value == nil || decoded.Value.Value != original.Value.Value {
		t.Fatalf("value: got %#v want %#v", decoded.Value, original.Value)
	}
	if decoded.Value.ValueKind != HostServiceCacheValueKindString {
		t.Fatalf("valueKind: got %d want %d", decoded.Value.ValueKind, HostServiceCacheValueKindString)
	}
}

func TestHostServiceCacheSetRequestRoundTrip(t *testing.T) {
	original := &HostServiceCacheSetRequest{
		Key:           "profile",
		Value:         `{"enabled":true}`,
		ExpireSeconds: 300,
	}

	data := MarshalHostServiceCacheSetRequest(original)
	decoded, err := UnmarshalHostServiceCacheSetRequest(data)
	if err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Key != original.Key {
		t.Fatalf("key: got %q want %q", decoded.Key, original.Key)
	}
	if decoded.Value != original.Value {
		t.Fatalf("value: got %q want %q", decoded.Value, original.Value)
	}
	if decoded.ExpireSeconds != original.ExpireSeconds {
		t.Fatalf("expireSeconds: got %d want %d", decoded.ExpireSeconds, original.ExpireSeconds)
	}
}

func TestHostServiceCacheIncrAndExpireRoundTrip(t *testing.T) {
	originalIncr := &HostServiceCacheIncrResponse{
		Value: &HostServiceCacheValue{
			ValueKind: HostServiceCacheValueKindInt,
			Value:     "12",
			IntValue:  12,
			ExpireAt:  "2026-04-15T12:00:01Z",
		},
	}
	incrData := MarshalHostServiceCacheIncrResponse(originalIncr)
	decodedIncr, err := UnmarshalHostServiceCacheIncrResponse(incrData)
	if err != nil {
		t.Fatalf("incr response unmarshal failed: %v", err)
	}
	if decodedIncr.Value == nil || decodedIncr.Value.IntValue != 12 {
		t.Fatalf("incr value: got %#v", decodedIncr.Value)
	}

	originalExpire := &HostServiceCacheExpireResponse{
		Found:    true,
		ExpireAt: "2026-04-15T12:00:02Z",
	}
	expireData := MarshalHostServiceCacheExpireResponse(originalExpire)
	decodedExpire, err := UnmarshalHostServiceCacheExpireResponse(expireData)
	if err != nil {
		t.Fatalf("expire response unmarshal failed: %v", err)
	}
	if !decodedExpire.Found || decodedExpire.ExpireAt != originalExpire.ExpireAt {
		t.Fatalf("expire response: got %#v want %#v", decodedExpire, originalExpire)
	}
}
