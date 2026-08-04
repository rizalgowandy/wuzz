package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfigCopiesDefaultKeys(t *testing.T) {
	tests := []struct {
		name       string
		contents   string
		customCtrl string
	}{
		{
			name: "no key configuration",
		},
		{
			name:       "omitted key category",
			contents:   "[keys.global]\nCtrlR = \"customSubmit\"\n",
			customCtrl: "customSubmit",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			original := DefaultKeys["url"]["Enter"]
			t.Cleanup(func() {
				DefaultKeys["url"]["Enter"] = original
			})

			configFile := filepath.Join(t.TempDir(), "config.toml")
			if err := os.WriteFile(configFile, []byte(test.contents), 0600); err != nil {
				t.Fatal(err)
			}

			loaded, err := LoadConfig(configFile)
			if err != nil {
				t.Fatal(err)
			}
			if test.customCtrl != "" && loaded.Keys["global"]["CtrlR"] != test.customCtrl {
				t.Fatalf("custom CtrlR binding = %q, want %q", loaded.Keys["global"]["CtrlR"], test.customCtrl)
			}

			loaded.Keys["url"]["Enter"] = "changed"
			if DefaultKeys["url"]["Enter"] != original {
				t.Fatalf("DefaultKeys changed to %q after loaded config mutation", DefaultKeys["url"]["Enter"])
			}

			reloaded, err := LoadConfig(configFile)
			if err != nil {
				t.Fatal(err)
			}
			if reloaded.Keys["url"]["Enter"] != original {
				t.Fatalf("subsequent load inherited %q, want %q", reloaded.Keys["url"]["Enter"], original)
			}
		})
	}
}
