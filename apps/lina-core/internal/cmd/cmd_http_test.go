package cmd

import "testing"

func TestParsePluginAssetRequestPath(t *testing.T) {
	tests := []struct {
		name          string
		path          string
		wantPluginID  string
		wantVersion   string
		wantAssetPath string
		wantOK        bool
	}{
		{
			name:          "hosted asset file",
			path:          "plugin-assets/plugin-demo-dynamic/v0.1.0/standalone.html",
			wantPluginID:  "plugin-demo-dynamic",
			wantVersion:   "v0.1.0",
			wantAssetPath: "standalone.html",
			wantOK:        true,
		},
		{
			name:          "embedded mount entry",
			path:          "/plugin-assets/plugin-demo-dynamic/v0.1.0/mount.js",
			wantPluginID:  "plugin-demo-dynamic",
			wantVersion:   "v0.1.0",
			wantAssetPath: "mount.js",
			wantOK:        true,
		},
		{
			name:          "version root path",
			path:          "/plugin-assets/plugin-demo-dynamic/v0.1.0/",
			wantPluginID:  "plugin-demo-dynamic",
			wantVersion:   "v0.1.0",
			wantAssetPath: "",
			wantOK:        true,
		},
		{
			name:   "non plugin path",
			path:   "/assets/index.js",
			wantOK: false,
		},
		{
			name:   "missing version",
			path:   "/plugin-assets/plugin-demo-dynamic",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotPluginID, gotVersion, gotAssetPath, gotOK := parsePluginAssetRequestPath(tt.path)
			if gotOK != tt.wantOK {
				t.Fatalf("expected ok=%v, got %v", tt.wantOK, gotOK)
			}
			if gotPluginID != tt.wantPluginID {
				t.Fatalf("expected pluginID=%q, got %q", tt.wantPluginID, gotPluginID)
			}
			if gotVersion != tt.wantVersion {
				t.Fatalf("expected version=%q, got %q", tt.wantVersion, gotVersion)
			}
			if gotAssetPath != tt.wantAssetPath {
				t.Fatalf("expected assetPath=%q, got %q", tt.wantAssetPath, gotAssetPath)
			}
		})
	}
}
