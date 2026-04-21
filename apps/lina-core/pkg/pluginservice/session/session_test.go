// This file verifies the published session seam keeps an independent DTO
// contract instead of leaking host-internal session types to plugins.

package session

import (
	"testing"

	"github.com/gogf/gf/v2/os/gtime"

	internalsession "lina-core/internal/service/session"
)

// TestToInternalFilter verifies the published filter contract is converted explicitly.
func TestToInternalFilter(t *testing.T) {
	if result := toInternalFilter(nil); result != nil {
		t.Fatalf("expected nil filter, got %#v", result)
	}

	result := toInternalFilter(&ListFilter{
		Username: "admin",
		Ip:       "127.0.0.1",
	})
	if result == nil {
		t.Fatal("expected converted filter, got nil")
	}
	if result.Username != "admin" || result.Ip != "127.0.0.1" {
		t.Fatalf("unexpected converted filter: %#v", result)
	}
}

// TestFromInternalSession verifies host-internal session projections are copied into the
// published DTO instead of being re-exported as type aliases.
func TestFromInternalSession(t *testing.T) {
	loginTime := gtime.Now()
	sessionItem := &internalsession.Session{
		TokenId:        "token-1",
		UserId:         100,
		Username:       "admin",
		DeptName:       "研发部",
		Ip:             "127.0.0.1",
		Browser:        "Chrome",
		Os:             "macOS",
		LoginTime:      loginTime,
		LastActiveTime: loginTime,
	}

	result := fromInternalSession(sessionItem)
	if result == nil {
		t.Fatal("expected converted session, got nil")
	}
	if result.TokenId != sessionItem.TokenId ||
		result.UserId != sessionItem.UserId ||
		result.Username != sessionItem.Username ||
		result.DeptName != sessionItem.DeptName ||
		result.Ip != sessionItem.Ip ||
		result.Browser != sessionItem.Browser ||
		result.Os != sessionItem.Os ||
		result.LoginTime != sessionItem.LoginTime ||
		result.LastActiveTime != sessionItem.LastActiveTime {
		t.Fatalf("unexpected converted session: %#v", result)
	}
}

// TestFromInternalListResult verifies nil-safe list conversion and item projection.
func TestFromInternalListResult(t *testing.T) {
	empty := fromInternalListResult(nil)
	if empty == nil {
		t.Fatal("expected empty result, got nil")
	}
	if empty.Total != 0 || len(empty.Items) != 0 {
		t.Fatalf("unexpected empty result: %#v", empty)
	}

	loginTime := gtime.Now()
	result := fromInternalListResult(&internalsession.ListResult{
		Items: []*internalsession.Session{
			{
				TokenId:        "token-2",
				UserId:         101,
				Username:       "demo",
				DeptName:       "测试部",
				Ip:             "10.0.0.1",
				Browser:        "Firefox",
				Os:             "Linux",
				LoginTime:      loginTime,
				LastActiveTime: loginTime,
			},
		},
		Total: 1,
	})
	if result.Total != 1 || len(result.Items) != 1 {
		t.Fatalf("unexpected converted list result: %#v", result)
	}
	if result.Items[0] == nil || result.Items[0].TokenId != "token-2" {
		t.Fatalf("unexpected converted item: %#v", result.Items[0])
	}
}
