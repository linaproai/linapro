// This file tests watermark strategy parsing.

package water

import "testing"

// TestParseWatermarkStrategyNested verifies Lina nested watermark YAML parsing.
func TestParseWatermarkStrategyNested(t *testing.T) {
	cfg, err := parseWatermarkStrategy(`record:
  enabled: true
watermark:
  enabled: true
  text: 园区安防
  fontSize: 48
  color: "#00ff88"
  align: bottomRight
  opacity: 0.5`)
	if err != nil {
		t.Fatalf("parse watermark strategy failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected watermark config")
	}
	if !cfg.Enabled {
		t.Fatal("expected enabled config")
	}
	if cfg.Text != "园区安防" {
		t.Fatalf("expected text to roundtrip, got %q", cfg.Text)
	}
	if cfg.FontSize != 48 {
		t.Fatalf("expected font size 48, got %d", cfg.FontSize)
	}
	if normalizedAlignment(cfg.Align) != "bottomright" {
		t.Fatalf("expected bottomright alignment, got %q", normalizedAlignment(cfg.Align))
	}
}

// TestParseWatermarkStrategyRoot verifies hotgo root-level compatibility.
func TestParseWatermarkStrategyRoot(t *testing.T) {
	cfg, err := parseWatermarkStrategy(`enabled: true
text: 热点水印
fontSize: 64
align: 9
opacity: 0.15`)
	if err != nil {
		t.Fatalf("parse watermark strategy failed: %v", err)
	}
	if cfg == nil {
		t.Fatal("expected root-level watermark config")
	}
	if normalizedAlignment(cfg.Align) != "bottomright" {
		t.Fatalf("expected numeric alignment to map to bottomright, got %q", normalizedAlignment(cfg.Align))
	}
	if cfg.Opacity != 0.15 {
		t.Fatalf("expected opacity 0.15, got %f", cfg.Opacity)
	}
}

// TestParseWatermarkStrategyMissing verifies strategies without watermark are skipped.
func TestParseWatermarkStrategyMissing(t *testing.T) {
	cfg, err := parseWatermarkStrategy(`record:
  enabled: true`)
	if err != nil {
		t.Fatalf("parse watermark strategy failed: %v", err)
	}
	if cfg != nil {
		t.Fatalf("expected nil watermark config, got %+v", cfg)
	}
}
