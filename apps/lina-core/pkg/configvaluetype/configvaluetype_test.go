// This file covers value-type resolution, options encoding, and typed value checks.

package configvaluetype

import "testing"

func TestResolveCode(t *testing.T) {
	t.Parallel()
	code, err := ResolveCode("")
	if err != nil || code != Text {
		t.Fatalf("empty should resolve to text, got %q err=%v", code, err)
	}
	code, err = ResolveCode("select")
	if err != nil || code != Select {
		t.Fatalf("select resolve failed: %q %v", code, err)
	}
	if _, err = ResolveCode("unknown"); err == nil {
		t.Fatal("expected unsupported type error")
	}
}

func TestValidateTypedValue(t *testing.T) {
	t.Parallel()
	options := `[{"label":"Left","value":"panel-left"},{"label":"Right","value":"panel-right"}]`
	if err := ValidateTypedValue(Boolean, "", "true"); err != nil {
		t.Fatalf("boolean true: %v", err)
	}
	if err := ValidateTypedValue(Boolean, "", "yes"); err == nil {
		t.Fatal("expected boolean rejection")
	}
	if err := ValidateTypedValue(Number, "", "12.5"); err != nil {
		t.Fatalf("number: %v", err)
	}
	if err := ValidateTypedValue(Number, "", "abc"); err == nil {
		t.Fatal("expected number rejection")
	}
	if err := ValidateTypedValue(Select, options, "panel-left"); err != nil {
		t.Fatalf("select: %v", err)
	}
	if err := ValidateTypedValue(Select, options, "panel-center"); err == nil {
		t.Fatal("expected select rejection")
	}
	if err := ValidateTypedValue(MultiSelect, options, "panel-left;panel-right"); err != nil {
		t.Fatalf("multi_select: %v", err)
	}
	if err := ValidateTypedValue(MultiSelect, options, "panel-left;bad"); err == nil {
		t.Fatal("expected multi_select rejection")
	}
	if err := ValidateTypedValue(Select, "", "x"); err == nil {
		t.Fatal("select requires options")
	}
}

func TestEncodeParseOptions(t *testing.T) {
	t.Parallel()
	raw, err := EncodeOptions([]Option{{Label: "A", Value: "a"}, {Label: "", Value: "b"}})
	if err != nil {
		t.Fatal(err)
	}
	options, err := ParseOptions(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(options) != 2 || options[1].Label != "b" {
		t.Fatalf("unexpected options: %+v", options)
	}
}

func TestParseOptionsSimpleLines(t *testing.T) {
	t.Parallel()
	options, err := ParseOptions("左侧=panel-left\n居中=panel-center\n\npanel-right\n")
	if err != nil {
		t.Fatal(err)
	}
	if len(options) != 3 {
		t.Fatalf("expected 3 options, got %+v", options)
	}
	if options[0].Label != "左侧" || options[0].Value != "panel-left" {
		t.Fatalf("unexpected first option: %+v", options[0])
	}
	if options[2].Label != "panel-right" || options[2].Value != "panel-right" {
		t.Fatalf("unexpected bare option: %+v", options[2])
	}
	pipeOptions, err := ParseOptions("Left|panel-left\nRight|panel-right")
	if err != nil {
		t.Fatal(err)
	}
	if len(pipeOptions) != 2 || pipeOptions[0].Value != "panel-left" {
		t.Fatalf("unexpected pipe options: %+v", pipeOptions)
	}
}

func TestFormatOptionsSimple(t *testing.T) {
	t.Parallel()
	got := FormatOptionsSimple([]Option{
		{Label: "左侧", Value: "panel-left"},
		{Label: "same", Value: "same"},
	})
	want := "左侧=panel-left\nsame"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestNormalizeMultiSelectValue(t *testing.T) {
	t.Parallel()
	got := NormalizeMultiSelectValue(" a ; b ; a ;;")
	if got != "a;b" {
		t.Fatalf("got %q", got)
	}
}
