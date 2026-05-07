// This file tests dialect factory behavior and MySQL public-contract behavior.

package dialect

import (
	"context"
	"testing"

	_ "github.com/gogf/gf/contrib/drivers/mysql/v2"
)

// TestFromResolvesSupportedDialects verifies link-prefix dispatch.
func TestFromResolvesSupportedDialects(t *testing.T) {
	tests := []struct {
		name        string
		link        string
		wantName    string
		wantCluster bool
	}{
		{
			name:        "mysql",
			link:        "mysql:root:pass@tcp(127.0.0.1:3306)/linapro",
			wantName:    "mysql",
			wantCluster: true,
		},
		{
			name:        "sqlite",
			link:        "sqlite::@file(./temp/sqlite/linapro.db)",
			wantName:    "sqlite",
			wantCluster: false,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			got, err := From(test.link)
			if err != nil {
				t.Fatalf("resolve dialect failed: %v", err)
			}
			if got.Name() != test.wantName {
				t.Fatalf("expected dialect %s, got %s", test.wantName, got.Name())
			}
			if got.SupportsCluster() != test.wantCluster {
				t.Fatalf("expected cluster support %t, got %t", test.wantCluster, got.SupportsCluster())
			}
		})
	}
}

// TestFromRejectsUnsupportedLink verifies unsupported prefixes do not silently
// fall back to a default dialect.
func TestFromRejectsUnsupportedLink(t *testing.T) {
	if _, err := From("postgres:postgres://localhost/linapro"); err == nil {
		t.Fatal("expected unsupported dialect to fail")
	}
}

// TestMySQLTranslateDDLNoop verifies MySQL translation preserves bytes.
func TestMySQLTranslateDDLNoop(t *testing.T) {
	ddl := "CREATE TABLE `demo` (`id` INT PRIMARY KEY AUTO_INCREMENT);"
	dbDialect, err := From("mysql:root:pass@tcp(127.0.0.1:3306)/linapro")
	if err != nil {
		t.Fatalf("resolve MySQL dialect failed: %v", err)
	}
	got, err := dbDialect.TranslateDDL(context.Background(), "fixture.sql", ddl)
	if err != nil {
		t.Fatalf("translate MySQL DDL failed: %v", err)
	}
	if got != ddl {
		t.Fatalf("expected MySQL DDL to be unchanged, got %q", got)
	}
}
