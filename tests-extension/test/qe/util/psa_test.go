package util

import "testing"

func TestSanitizePSAPath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{name: "namespace", input: "ns/example-123", want: "ns-example-123"},
		{name: "empty", input: "", want: "unknown"},
		{name: "safe characters", input: "namespace_1.test", want: "namespace_1.test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := sanitizePSAPath(tt.input); got != tt.want {
				t.Fatalf("sanitizePSAPath(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
