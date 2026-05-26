package validate

import "testing"

func TestTarget(t *testing.T) {
	tests := []struct {
		name string
		ip   string
		port int
		ok   bool
	}{
		{name: "private 192", ip: "192.168.1.10", port: 39000, ok: true},
		{name: "private 10", ip: "10.0.0.8", port: 39000, ok: true},
		{name: "public", ip: "8.8.8.8", port: 39000},
		{name: "loopback", ip: "127.0.0.1", port: 39000},
		{name: "bad port", ip: "192.168.1.10", port: 70000},
		{name: "bad ip", ip: "hello", port: 39000},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Target(tt.ip, tt.port)
			if tt.ok && err != nil {
				t.Fatalf("expected ok, got %v", err)
			}
			if !tt.ok && err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
