// This file verifies LinaPro's shared GoFrame database driver registry.

package dbdriver

import "testing"

// TestSupportedTypesReturnsCopy verifies callers cannot mutate the package
// driver registry through the returned slice.
func TestSupportedTypesReturnsCopy(t *testing.T) {
	types := SupportedTypes()
	if len(types) != 1 {
		t.Fatalf("expected one supported database driver, got %d", len(types))
	}
	if types[0] != TypePostgreSQL {
		t.Fatalf("unexpected supported driver order: %#v", types)
	}

	types[0] = "mutated"
	next := SupportedTypes()
	if next[0] != TypePostgreSQL {
		t.Fatalf("expected supported driver registry to be immutable, got %#v", next)
	}
}

// TestNewCreatesSupportedDrivers verifies the shared factory handles supported
// driver names and rejects unsupported names.
func TestNewCreatesSupportedDrivers(t *testing.T) {
	for _, driverType := range []string{" pgsql ", "PGSQL"} {
		driver, ok := New(driverType)
		if !ok {
			t.Fatalf("expected driver type %q to be supported", driverType)
		}
		if driver == nil {
			t.Fatalf("expected driver type %q to create a non-nil driver", driverType)
		}
	}

	if IsSupported("mysql") {
		t.Fatal("expected mysql driver type to be unsupported")
	}
	if IsSupported("sqlite") {
		t.Fatal("expected sqlite driver type to be unsupported")
	}
	if driver, ok := New("mysql"); ok || driver != nil {
		t.Fatalf("expected unsupported mysql driver to return nil/false, got driver=%v ok=%t", driver, ok)
	}
	if driver, ok := New("sqlite"); ok || driver != nil {
		t.Fatalf("expected unsupported sqlite driver to return nil/false, got driver=%v ok=%t", driver, ok)
	}
}
