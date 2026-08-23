package normalize

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

func decodeJSONValue(raw json.RawMessage) (any, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	return value, nil
}

func fingerprintRaw(raw json.RawMessage) (string, []string, error) {
	value, err := decodeJSONValue(raw)
	if err != nil {
		return "", nil, err
	}
	keys := topLevelKeys(value)
	shape := structuralShape(value)
	hash := sha256.Sum256([]byte(shape))
	return hex.EncodeToString(hash[:]), keys, nil
}

func fingerprintType(raw json.RawMessage) string {
	if len(raw) == 0 {
		return hashString("unknown")
	}
	value, err := decodeJSONValue(raw)
	if err != nil {
		return hashString("unknown")
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return hashString("unknown")
	}
	return hashString(string(encoded))
}

func hashString(value string) string {
	hash := sha256.Sum256([]byte(value))
	return hex.EncodeToString(hash[:])
}

func topLevelKeys(value any) []string {
	object, ok := value.(map[string]any)
	if !ok {
		return []string{}
	}
	keys := make([]string, 0, len(object))
	for key := range object {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func structuralShape(value any) string {
	switch typed := value.(type) {
	case nil:
		return "null"
	case bool:
		return "bool"
	case json.Number, float64:
		return "number"
	case string:
		return "string"
	case []any:
		shapes := make(map[string]struct{})
		for _, item := range typed {
			shapes[structuralShape(item)] = struct{}{}
		}
		items := make([]string, 0, len(shapes))
		for shape := range shapes {
			items = append(items, shape)
		}
		sort.Strings(items)
		return "array[" + strings.Join(items, "|") + "]"
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			parts = append(parts, fmt.Sprintf("%s:%s", key, structuralShape(typed[key])))
		}
		return "object{" + strings.Join(parts, ",") + "}"
	default:
		return fmt.Sprintf("unknown:%T", typed)
	}
}

func markerPaths(raw json.RawMessage) []string {
	value, err := decodeJSONValue(raw)
	if err != nil || value == nil {
		return []string{}
	}
	paths := make([]string, 0)
	collectMarkerPaths(value, "", &paths)
	sort.Strings(paths)
	return uniqueStrings(paths)
}

func collectMarkerPaths(value any, path string, paths *[]string) {
	switch typed := value.(type) {
	case bool:
		if typed && path != "" {
			*paths = append(*paths, path)
		}
	case map[string]any:
		keys := make([]string, 0, len(typed))
		for key := range typed {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		for _, key := range keys {
			collectMarkerPaths(typed[key], joinPath(path, key), paths)
		}
	case []any:
		for index, item := range typed {
			collectMarkerPaths(item, fmt.Sprintf("%s[%d]", path, index), paths)
		}
	}
}

func joinPath(parent, child string) string {
	if parent == "" {
		return child
	}
	return parent + "." + child
}

func uniqueStrings(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	result := values[:0]
	for index, value := range values {
		if index == 0 || value != values[index-1] {
			result = append(result, value)
		}
	}
	return result
}
