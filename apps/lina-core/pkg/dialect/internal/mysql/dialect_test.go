// This file tests MySQL dialect configuration parsing and identifier quoting.

package mysql

import (
	"strings"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

// TestConfigNodeFromLink verifies database links are parsed by GoFrame's
// database component instead of package-local string splitting.
func TestConfigNodeFromLink(t *testing.T) {
	link := "mysql:root:12345678@tcp(127.0.0.1:3306)/linapro?charset=utf8mb4&parseTime=true&loc=Local"
	node, err := ConfigNodeFromLink(link)
	if err != nil {
		t.Fatalf("parse MySQL config node failed: %v", err)
	}
	if node.Type != "mysql" {
		t.Fatalf("expected mysql type, got %q", node.Type)
	}
	if node.User != "root" {
		t.Fatalf("expected root user, got %q", node.User)
	}
	if node.Pass != "12345678" {
		t.Fatalf("expected password to be preserved, got %q", node.Pass)
	}
	if node.Protocol != "tcp" {
		t.Fatalf("expected tcp protocol, got %q", node.Protocol)
	}
	if node.Host != "127.0.0.1" {
		t.Fatalf("expected localhost host, got %q", node.Host)
	}
	if node.Port != "3306" {
		t.Fatalf("expected 3306 port, got %q", node.Port)
	}
	if node.Name != "linapro" {
		t.Fatalf("expected linapro database name, got %q", node.Name)
	}
	if node.Charset != "utf8mb4" {
		t.Fatalf("expected utf8mb4 charset, got %q", node.Charset)
	}
	if !strings.Contains(node.Extra, "parseTime=true") || !strings.Contains(node.Extra, "loc=Local") {
		t.Fatalf("expected extra parameters to be preserved, got %q", node.Extra)
	}
}

// TestConfigNodeFromLinkUsesConfiguredDatabase verifies initialization derives
// the target schema from the configured link instead of a fixed name.
func TestConfigNodeFromLinkUsesConfiguredDatabase(t *testing.T) {
	link := "mysql:root:12345678@tcp(127.0.0.1:3306)/custom_app?charset=utf8mb4"
	node, err := ConfigNodeFromLink(link)
	if err != nil {
		t.Fatalf("parse configured database link failed: %v", err)
	}
	if node.Name != "custom_app" {
		t.Fatalf("expected configured database name custom_app, got %q", node.Name)
	}

	quotedName, err := QuoteIdentifier(node.Name)
	if err != nil {
		t.Fatalf("quote configured database name failed: %v", err)
	}
	if quotedName != "`custom_app`" {
		t.Fatalf("expected quoted configured database name, got %s", quotedName)
	}
}

// TestConfigNodeFromLinkRejectsMissingDatabase verifies init refuses links that
// cannot identify the target schema to create or rebuild.
func TestConfigNodeFromLinkRejectsMissingDatabase(t *testing.T) {
	for _, link := range []string{
		"mysql:root:12345678@tcp(127.0.0.1:3306)",
		"mysql:root:12345678@tcp(127.0.0.1:3306)/?charset=utf8mb4",
	} {
		link := link
		t.Run(link, func(t *testing.T) {
			if _, err := ConfigNodeFromLink(link); err == nil {
				t.Fatal("expected missing database name error")
			}
		})
	}
}

// TestQuoteIdentifier verifies bootstrap SQL escapes MySQL identifiers instead
// of concatenating raw names.
func TestQuoteIdentifier(t *testing.T) {
	got, err := QuoteIdentifier("lina`pro")
	if err != nil {
		t.Fatalf("quote identifier: %v", err)
	}
	if got != "`lina``pro`" {
		t.Fatalf("expected escaped identifier, got %q", got)
	}
	if _, err = QuoteIdentifier(""); err == nil {
		t.Fatal("expected empty identifier error")
	}
}
