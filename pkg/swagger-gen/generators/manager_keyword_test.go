package generators

import (
	"bytes"
	"os"
	"os/exec"
	"testing"
)

func TestGetManagerKeyword(t *testing.T) {
	for name, want := range map[string]string{
		"HostManager":         "host",
		"SHostManager":        "host",
		"SHostNetworkManager": "host_network",
		"NetworkManager":      "network",
	} {
		t.Run(name, func(t *testing.T) {
			if got := getManagerKeyword(name); got != want {
				t.Fatalf("getManagerKeyword(%q) = %q, want %q", name, got, want)
			}
		})
	}
}

func TestGetManagerKeywordRejectsShortNames(t *testing.T) {
	const helperEnv = "CODEGEN_TEST_MANAGER_NAME"
	if name, helper := os.LookupEnv(helperEnv); helper {
		getManagerKeyword(name)
		return
	}
	for _, name := range []string{"", "Manager", "S", "SManager", "A", "AManager"} {
		t.Run(name, func(t *testing.T) {
			cmd := exec.Command(os.Args[0], "-test.run=^TestGetManagerKeywordRejectsShortNames$")
			cmd.Env = append(os.Environ(), helperEnv+"="+name)
			output, err := cmd.CombinedOutput()
			if _, ok := err.(*exec.ExitError); !ok {
				t.Fatalf("expected rejection for %q, got %v: %s", name, err, output)
			}
			if !bytes.Contains(output, []byte("Invalid manager struct name:")) {
				t.Fatalf("expected validation error for %q, got: %s", name, output)
			}
		})
	}
}
