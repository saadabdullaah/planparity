package runner

import (
	"os"
	"strings"
	"testing"
)

func TestSafeEnvironmentStripsCredentials(t *testing.T) {
	for name, value := range map[string]string{
		"AWS_ACCESS_KEY_ID": "canary-aws",
		"GITHUB_TOKEN":      "canary-github",
		"MY_PASSWORD":       "canary-password",
		"PATH":              os.Getenv("PATH"),
		"HTTPS_PROXY":       "http://proxy.invalid",
	} {
		t.Setenv(name, value)
	}
	environment := strings.Join(SafeEnvironment(EnvironmentOptions{InheritProxy: false, TemporaryHome: "C:/temp/home", DataDirectory: "C:/temp/data"}), "\n")
	for _, canary := range []string{"canary-aws", "canary-github", "canary-password", "proxy.invalid"} {
		if strings.Contains(environment, canary) {
			t.Fatalf("credential or proxy escaped: %s", canary)
		}
	}
	for _, expected := range []string{"HOME=C:/temp/home", "TF_DATA_DIR=C:/temp/data", "TF_IN_AUTOMATION=1"} {
		if !strings.Contains(environment, expected) {
			t.Fatalf("missing %s in %s", expected, environment)
		}
	}
}
