// This file tests transport codecs shared by user, org, and tenant capability host services.

package protocol

import (
	"reflect"
	"testing"
)

// TestHostServiceCapabilityCodecsRoundTrip verifies primitive request and JSON
// response codecs used by user, organization, and tenant host services.
func TestHostServiceCapabilityCodecsRoundTrip(t *testing.T) {
	userBatchRequest := &HostServiceUsersBatchGetRequest{UserIDs: []string{"42", "platform:admin"}}
	decodedUserBatchRequest, err := UnmarshalHostServiceUsersBatchGetRequest(
		MarshalHostServiceUsersBatchGetRequest(userBatchRequest),
	)
	if err != nil {
		t.Fatalf("decode user batch request: %v", err)
	}
	if !reflect.DeepEqual(decodedUserBatchRequest, userBatchRequest) {
		t.Fatalf("unexpected decoded user batch request: %#v", decodedUserBatchRequest)
	}

	userSearchRequest := &HostServiceUsersSearchRequest{Keyword: "admin", PageNum: 2, PageSize: 20}
	decodedUserSearchRequest, err := UnmarshalHostServiceUsersSearchRequest(
		MarshalHostServiceUsersSearchRequest(userSearchRequest),
	)
	if err != nil {
		t.Fatalf("decode user search request: %v", err)
	}
	if !reflect.DeepEqual(decodedUserSearchRequest, userSearchRequest) {
		t.Fatalf("unexpected decoded user search request: %#v", decodedUserSearchRequest)
	}

	userEnsureRequest := &HostServiceUsersEnsureVisibleRequest{UserIDs: []string{"7", "8"}}
	decodedUserEnsureRequest, err := UnmarshalHostServiceUsersEnsureVisibleRequest(
		MarshalHostServiceUsersEnsureVisibleRequest(userEnsureRequest),
	)
	if err != nil {
		t.Fatalf("decode user ensure request: %v", err)
	}
	if !reflect.DeepEqual(decodedUserEnsureRequest, userEnsureRequest) {
		t.Fatalf("unexpected decoded user ensure request: %#v", decodedUserEnsureRequest)
	}

	userRequest := &HostServiceCapabilityUserRequest{UserID: 42}
	decodedUserRequest, err := UnmarshalHostServiceCapabilityUserRequest(
		MarshalHostServiceCapabilityUserRequest(userRequest),
	)
	if err != nil {
		t.Fatalf("decode user request failed: %v", err)
	}
	if !reflect.DeepEqual(decodedUserRequest, userRequest) {
		t.Fatalf("unexpected user request: %#v", decodedUserRequest)
	}

	usersRequest := &HostServiceCapabilityUsersRequest{UserIDs: []int{7, 8}}
	decodedUsersRequest, err := UnmarshalHostServiceCapabilityUsersRequest(
		MarshalHostServiceCapabilityUsersRequest(usersRequest),
	)
	if err != nil {
		t.Fatalf("decode users request failed: %v", err)
	}
	if !reflect.DeepEqual(decodedUsersRequest, usersRequest) {
		t.Fatalf("unexpected users request: %#v", decodedUsersRequest)
	}

	userTenantRequest := &HostServiceCapabilityUserTenantRequest{UserID: 42, TenantID: 3}
	decodedUserTenantRequest, err := UnmarshalHostServiceCapabilityUserTenantRequest(
		MarshalHostServiceCapabilityUserTenantRequest(userTenantRequest),
	)
	if err != nil {
		t.Fatalf("decode user tenant request failed: %v", err)
	}
	if !reflect.DeepEqual(decodedUserTenantRequest, userTenantRequest) {
		t.Fatalf("unexpected user tenant request: %#v", decodedUserTenantRequest)
	}

	switchRequest := &HostServiceCapabilityUserTenantSwitchRequest{UserID: 42, TargetTenantID: 3}
	decodedSwitchRequest, err := UnmarshalHostServiceCapabilityUserTenantSwitchRequest(
		MarshalHostServiceCapabilityUserTenantSwitchRequest(switchRequest),
	)
	if err != nil {
		t.Fatalf("decode tenant switch request failed: %v", err)
	}
	if !reflect.DeepEqual(decodedSwitchRequest, switchRequest) {
		t.Fatalf("unexpected tenant switch request: %#v", decodedSwitchRequest)
	}

	response := &HostServiceCapabilityJSONResponse{Value: []byte(`{"ok":true}`)}
	decodedResponse, err := UnmarshalHostServiceCapabilityJSONResponse(
		MarshalHostServiceCapabilityJSONResponse(response),
	)
	if err != nil {
		t.Fatalf("decode JSON response failed: %v", err)
	}
	if !reflect.DeepEqual(decodedResponse, response) {
		t.Fatalf("unexpected JSON response: %#v", decodedResponse)
	}
}
