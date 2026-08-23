package command

import (
	"reflect"
	"testing"
)

func TestBrowserCommandUsesAnArgumentVector(t *testing.T) {
	t.Parallel()
	url := "http://127.0.0.1:8765/#token=value&next=calc.exe"
	tests := []struct {
		goos string
		want []string
	}{
		{goos: "windows", want: []string{"rundll32", "url.dll,FileProtocolHandler", url}},
		{goos: "linux", want: []string{"xdg-open", url}},
	}
	for _, test := range tests {
		t.Run(test.goos, func(t *testing.T) {
			command, err := browserCommand(test.goos, url)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(command.Args, test.want) {
				t.Fatalf("args = %#v, want %#v", command.Args, test.want)
			}
		})
	}
}

func TestBrowserCommandRejectsUnsupportedPlatforms(t *testing.T) {
	t.Parallel()
	if _, err := browserCommand("plan9", "http://127.0.0.1"); err == nil {
		t.Fatal("unsupported platform was accepted")
	}
}
