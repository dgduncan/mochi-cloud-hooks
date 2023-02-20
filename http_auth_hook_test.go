package mochicloudhooks

import (
	"testing"

	"github.com/mochi-co/mqtt/v2"
	"github.com/stretchr/testify/require"
)

func TestID(t *testing.T) {
	authHook := new(HTTPAuthHook)

	require.Equal(t, "http-auth-hook", authHook.ID())
}

func TestProvides(t *testing.T) {
	authHook := new(HTTPAuthHook)

	tests := []struct {
		name           string
		hook           byte
		expectProvides bool
	}{
		{
			name:           "Success - Provides OnACLCheck",
			hook:           mqtt.OnACLCheck,
			expectProvides: true,
		},
		{
			name:           "Success - Provides OnConnectAuthenticate",
			hook:           mqtt.OnConnectAuthenticate,
			expectProvides: true,
		},
		{
			name:           "Failure - Provides other hook",
			hook:           mqtt.OnClientExpired,
			expectProvides: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			require.Equal(t, tt.expectProvides, authHook.Provides(tt.hook))

		})
	}
}
