package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// A container runtime tells the application where to listen through the
// environment; a hardcoded address is unreachable from outside the container
// and ignores the port that was actually published.
func TestListenAddr(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want string
	}{
		{"binds all interfaces by default", nil, "0.0.0.0:8585"},
		{"honours PORT", map[string]string{"PORT": "9999"}, "0.0.0.0:9999"},
		{"honours HOST", map[string]string{"HOST": "127.0.0.1"}, "127.0.0.1:8585"},
		{"honours both", map[string]string{"HOST": "127.0.0.1", "PORT": "7777"}, "127.0.0.1:7777"},
		{"ADDR wins outright", map[string]string{"ADDR": "10.0.0.1:1234", "PORT": "9999"}, "10.0.0.1:1234"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for k, v := range tc.env {
				t.Setenv(k, v)
			}
			assert.Equal(t, tc.want, listenAddr())
		})
	}
}
