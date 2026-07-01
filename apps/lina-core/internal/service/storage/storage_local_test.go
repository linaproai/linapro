// This file verifies the host-internal local object storage backend.

package storage

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestLocalStoragePutGetStatDelete(t *testing.T) {
	var (
		ctx     = context.Background()
		root    = t.TempDir()
		service = New(Config{RootDir: root})
	)

	putOutput, err := service.Put(ctx, PutInput{
		Namespace:   NamespaceFiles,
		Key:         "42/2026/06/demo.txt",
		Body:        strings.NewReader("demo"),
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("put object: %v", err)
	}
	if putOutput == nil || putOutput.Object == nil || putOutput.Object.Key != "42/2026/06/demo.txt" {
		t.Fatalf("unexpected put output: %#v", putOutput)
	}

	content, err := os.ReadFile(filepath.Join(root, "42", "2026", "06", "demo.txt"))
	if err != nil {
		t.Fatalf("read stored object from disk: %v", err)
	}
	if string(content) != "demo" {
		t.Fatalf("expected stored content demo, got %q", string(content))
	}

	getOutput, err := service.Get(ctx, GetInput{Namespace: NamespaceFiles, Key: "42/2026/06/demo.txt"})
	if err != nil {
		t.Fatalf("get object: %v", err)
	}
	if getOutput == nil || !getOutput.Found || getOutput.Body == nil {
		t.Fatalf("expected found object, got %#v", getOutput)
	}
	body, err := io.ReadAll(getOutput.Body)
	if err != nil {
		t.Fatalf("read object body: %v", err)
	}
	if closeErr := getOutput.Body.Close(); closeErr != nil {
		t.Fatalf("close object body: %v", closeErr)
	}
	if string(body) != "demo" {
		t.Fatalf("expected read content demo, got %q", string(body))
	}

	statOutput, err := service.Stat(ctx, StatInput{Namespace: NamespaceFiles, Key: "42/2026/06/demo.txt"})
	if err != nil {
		t.Fatalf("stat object: %v", err)
	}
	if statOutput == nil || !statOutput.Found || statOutput.Object.Size != int64(len("demo")) || statOutput.Object.ETag == "" {
		t.Fatalf("unexpected stat output: %#v", statOutput)
	}

	if err = service.Delete(ctx, DeleteInput{Namespace: NamespaceFiles, Key: "42/2026/06/demo.txt"}); err != nil {
		t.Fatalf("delete object: %v", err)
	}
	missing, err := service.Get(ctx, GetInput{Namespace: NamespaceFiles, Key: "42/2026/06/demo.txt"})
	if err != nil {
		t.Fatalf("get deleted object: %v", err)
	}
	if missing == nil || missing.Found {
		t.Fatalf("expected deleted object missing, got %#v", missing)
	}
}

func TestLocalStorageRejectsUnsafePaths(t *testing.T) {
	ctx := context.Background()
	service := New(Config{RootDir: t.TempDir()})
	for _, key := range []string{"", ".", "../secret.txt", "/secret.txt", "a/../../secret.txt", "C:/secret.txt", "https://example.test/a.txt"} {
		t.Run(key, func(t *testing.T) {
			_, err := service.Put(ctx, PutInput{Namespace: NamespaceFiles, Key: key, Body: strings.NewReader("x")})
			if !errors.Is(err, ErrPathInvalid) {
				t.Fatalf("expected ErrPathInvalid for %q, got %v", key, err)
			}
		})
	}
}

func TestLocalStorageUsesNamespaceRoots(t *testing.T) {
	var (
		ctx        = context.Background()
		fileRoot   = t.TempDir()
		pluginRoot = t.TempDir()
	)
	service := New(Config{NamespaceRoots: map[string]string{
		NamespaceFiles:   fileRoot,
		NamespacePlugins: pluginRoot,
	}})

	if _, err := service.Put(ctx, PutInput{Namespace: NamespaceFiles, Key: "same/key.txt", Body: strings.NewReader("file")}); err != nil {
		t.Fatalf("put file object: %v", err)
	}
	if _, err := service.Put(ctx, PutInput{Namespace: NamespacePlugins, Key: "same/key.txt", Body: strings.NewReader("plugin")}); err != nil {
		t.Fatalf("put plugin object: %v", err)
	}
	fileContent, err := os.ReadFile(filepath.Join(fileRoot, "same", "key.txt"))
	if err != nil {
		t.Fatalf("read file namespace object: %v", err)
	}
	pluginContent, err := os.ReadFile(filepath.Join(pluginRoot, "same", "key.txt"))
	if err != nil {
		t.Fatalf("read plugin namespace object: %v", err)
	}
	if string(fileContent) != "file" || string(pluginContent) != "plugin" {
		t.Fatalf("unexpected namespace contents file=%q plugin=%q", string(fileContent), string(pluginContent))
	}
}

func TestLocalStorageListsPrefixWithLimitAndCursor(t *testing.T) {
	ctx := context.Background()
	service := New(Config{RootDir: t.TempDir()})
	for _, key := range []string{
		"plugins/reporting/platform/reports/a.json",
		"plugins/reporting/platform/reports/b.json",
		"plugins/reporting/platform/reports/c.json",
		"plugins/reporting/platform/private/hidden.json",
	} {
		if _, err := service.Put(ctx, PutInput{Namespace: NamespacePlugins, Key: key, Body: strings.NewReader(key)}); err != nil {
			t.Fatalf("put object %s: %v", key, err)
		}
	}

	listOutput, err := service.List(ctx, ListInput{
		Namespace: NamespacePlugins,
		Prefix:    "plugins/reporting/platform/reports",
		Limit:     2,
	})
	if err != nil {
		t.Fatalf("list objects: %v", err)
	}
	got := objectKeys(listOutput.Objects)
	want := []string{
		"plugins/reporting/platform/reports/a.json",
		"plugins/reporting/platform/reports/b.json",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("expected list keys %#v, got %#v", want, got)
	}

	cursorOutput, err := service.ListCursor(ctx, ListCursorInput{
		Namespace: NamespacePlugins,
		Prefix:    "plugins/reporting/platform/reports",
		Cursor:    "plugins/reporting/platform/reports/a.json",
		Limit:     1,
	})
	if err != nil {
		t.Fatalf("list cursor objects: %v", err)
	}
	got = objectKeys(cursorOutput.Objects)
	want = []string{"plugins/reporting/platform/reports/b.json"}
	if !reflect.DeepEqual(got, want) || cursorOutput.NextCursor == "" {
		t.Fatalf("expected cursor page %#v with next cursor, got keys=%#v cursor=%q", want, got, cursorOutput.NextCursor)
	}
}

func objectKeys(objects []*Object) []string {
	keys := make([]string, 0, len(objects))
	for _, object := range objects {
		if object != nil {
			keys = append(keys, object.Key)
		}
	}
	return keys
}
