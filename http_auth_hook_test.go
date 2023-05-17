package mochicloudhooks

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	gomock "github.com/golang/mock/gomock"
	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
)

var defaultClientID = "default_client_id"

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

func TestInit(t *testing.T) {
	authHook := new(HTTPAuthHook)
	authHook.Log = &zerolog.Logger{}

	tests := []struct {
		name        string
		config      any
		expectError bool
	}{
		{
			name:        "Success - Proper config",
			config:      HTTPAuthHookConfig{},
			expectError: false,
		},
		{
			name:        "Failure - nil config",
			config:      nil,
			expectError: true,
		},
		{
			name:        "Failure - improper config",
			config:      "",
			expectError: true,
		},
		{
			name:        "Failure - hostname validation fail",
			config:      HTTPAuthHookConfig{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			err := authHook.Init(tt.config)
			if tt.expectError {
				require.Error(t, err)
			}

		})
	}
}

func TestOnACLCheck(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)

	tests := []struct {
		name           string
		config         any
		clientBlockMap map[string]time.Time
		mocks          func(ctx context.Context)
		expectPass     bool
	}{
		{
			name: "Success - Proper config",
			config: HTTPAuthHookConfig{
				RoundTripper:             mockRT,
				ACLHost:                  "http://aclhost.com",
				ClientAuthenticationHost: "http://clientauthenticationhost.com",
			},
			expectPass: true,
			mocks: func(ctx context.Context) {
				mockRT.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
					StatusCode: http.StatusOK,
				}, nil)

			},
		},
		{
			name: "Success - Proper config - Timeout Configured - Not Blocked",
			config: HTTPAuthHookConfig{
				RoundTripper:             mockRT,
				ACLHost:                  "http://aclhost.com",
				ClientAuthenticationHost: "http://clientauthenticationhost.com",
				Timeout: TimeoutConfig{
					TimeoutDuration: 1 * time.Microsecond,
				},
			},
			expectPass: true,
			mocks: func(ctx context.Context) {
				mockRT.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
					StatusCode: http.StatusOK,
				}, nil)

			},
		},
		{
			name: "Success - Proper config - Timeout Configured - Already Blocked",
			config: HTTPAuthHookConfig{
				RoundTripper:             mockRT,
				ACLHost:                  "http://aclhost.com",
				ClientAuthenticationHost: "http://clientauthenticationhost.com",
				Timeout: TimeoutConfig{
					TimeoutDuration: 1 * time.Microsecond,
				},
			},
			clientBlockMap: map[string]time.Time{
				"defaultClientID": time.Now().Add(1 * time.Minute),
			},
			expectPass: false,
			mocks: func(ctx context.Context) {
			},
		},
		{
			name: "Success - Proper config - Timeout Configured - Should Block",
			config: HTTPAuthHookConfig{
				RoundTripper:             mockRT,
				ACLHost:                  "http://aclhost.com",
				ClientAuthenticationHost: "http://clientauthenticationhost.com",
				Timeout: TimeoutConfig{
					TimeoutDuration: 1 * time.Microsecond,
				},
			},
			expectPass: false,
			mocks: func(ctx context.Context) {
				mockRT.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
					StatusCode: http.StatusUnauthorized,
				}, nil)

			},
		},
		{
			name: "Success - Proper config - Timeout Configured - Should Delete Client From Block Map",
			config: HTTPAuthHookConfig{
				RoundTripper:             mockRT,
				ACLHost:                  "http://aclhost.com",
				ClientAuthenticationHost: "http://clientauthenticationhost.com",
				Timeout: TimeoutConfig{
					TimeoutDuration: 1 * time.Microsecond,
				},
			},
			clientBlockMap: map[string]time.Time{
				"defaultClientID": time.Now().Add(-1 * time.Hour),
			},
			expectPass: true,
			mocks: func(ctx context.Context) {
				mockRT.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
					StatusCode: http.StatusOK,
				}, nil)

			},
		},
		{
			name: "Error - HTTP error",
			config: HTTPAuthHookConfig{
				RoundTripper:             mockRT,
				ACLHost:                  "http://aclhost.com",
				ClientAuthenticationHost: "http://clientauthenticationhost.com",
			},
			expectPass: false,
			mocks: func(ctx context.Context) {
				mockRT.EXPECT().RoundTrip(gomock.Any()).Return(nil, errors.New("Oh Crap"))
			},
		},
		{
			name: "Error - Non 2xx",
			config: HTTPAuthHookConfig{
				RoundTripper:             mockRT,
				ACLHost:                  "http://aclhost.com",
				ClientAuthenticationHost: "http://clientauthenticationhost.com",
			},
			expectPass: false,
			mocks: func(ctx context.Context) {
				mockRT.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
					StatusCode: http.StatusTeapot,
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tt.mocks(ctx)

			authHook := new(HTTPAuthHook)
			authHook.Log = &zerolog.Logger{}
			authHook.Init(tt.config)

			if tt.clientBlockMap != nil {
				authHook.clientBlockMap = tt.clientBlockMap
			}

			success := authHook.OnACLCheck(&mqtt.Client{
				ID: "defaultClientID",
			}, "/topic", false)

			require.Equal(t, tt.expectPass, success)
		})
	}
}

func TestOnConnectAuthenticate(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRT := NewMockRoundTripper(ctrl)

	tests := []struct {
		name   string
		config any
		mocks  func(ctx context.Context)

		expectPass bool
	}{
		{
			name: "Success - Proper config",
			config: HTTPAuthHookConfig{
				RoundTripper:             mockRT,
				ACLHost:                  "http://aclhost.com",
				ClientAuthenticationHost: "http://clientauthenticationhost.com",
			},
			expectPass: true,
			mocks: func(ctx context.Context) {
				mockRT.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
					StatusCode: http.StatusOK,
				}, nil)

			},
		},
		{
			name: "Error - HTTP error",
			config: HTTPAuthHookConfig{
				RoundTripper:             mockRT,
				ACLHost:                  "http://aclhost.com",
				ClientAuthenticationHost: "http://clientauthenticationhost.com",
			},
			expectPass: false,
			mocks: func(ctx context.Context) {
				mockRT.EXPECT().RoundTrip(gomock.Any()).Return(nil, errors.New("Oh Crap"))
			},
		},
		{
			name: "Error - Non 2xx",
			config: HTTPAuthHookConfig{
				RoundTripper:             mockRT,
				ACLHost:                  "http://aclhost.com",
				ClientAuthenticationHost: "http://clientauthenticationhost.com",
			},
			expectPass: false,
			mocks: func(ctx context.Context) {
				mockRT.EXPECT().RoundTrip(gomock.Any()).Return(&http.Response{
					StatusCode: http.StatusTeapot,
				}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			tt.mocks(ctx)

			authHook := new(HTTPAuthHook)
			authHook.Log = &zerolog.Logger{}
			authHook.Init(tt.config)

			success := authHook.OnConnectAuthenticate(&mqtt.Client{}, packets.Packet{})
			require.Equal(t, tt.expectPass, success)
		})
	}
}
