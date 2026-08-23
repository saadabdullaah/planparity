package runner

import (
	"os"
	"sort"
	"strings"
)

// EnvironmentOptions controls the small, documented inheritance surface.
type EnvironmentOptions struct {
	InheritProxy  bool
	TemporaryHome string
	DataDirectory string
}

// SafeEnvironment creates an allowlisted environment and strips credential-like names.
func SafeEnvironment(options EnvironmentOptions) []string {
	allowed := map[string]bool{
		"PATH": true, "PATHEXT": true, "SYSTEMROOT": true, "WINDIR": true,
		"COMSPEC": true, "TEMP": true, "TMP": true, "TMPDIR": true,
		"TF_PLUGIN_CACHE_DIR": true,
	}
	if options.InheritProxy {
		for _, name := range []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "ALL_PROXY"} {
			allowed[name] = true
		}
	}
	values := make(map[string]string)
	for _, entry := range os.Environ() {
		name, value, found := strings.Cut(entry, "=")
		if !found {
			continue
		}
		upper := strings.ToUpper(name)
		if !allowed[upper] || secretLikeName(upper) {
			continue
		}
		values[upper] = value
	}
	if options.TemporaryHome != "" {
		values["HOME"] = options.TemporaryHome
		values["USERPROFILE"] = options.TemporaryHome
	}
	if options.DataDirectory != "" {
		values["TF_DATA_DIR"] = options.DataDirectory
	}
	values["TF_IN_AUTOMATION"] = "1"
	values["CHECKPOINT_DISABLE"] = "1"
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]string, 0, len(keys))
	for _, key := range keys {
		result = append(result, key+"="+values[key])
	}
	return result
}

func secretLikeName(name string) bool {
	for _, fragment := range []string{"TOKEN", "SECRET", "PASSWORD", "CREDENTIAL"} {
		if strings.Contains(name, fragment) {
			return true
		}
	}
	for _, prefix := range []string{"AWS_", "ARM_", "AZURE_", "GOOGLE_", "CLOUDSDK_", "VAULT_", "GITHUB_", "TF_TOKEN_"} {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}
