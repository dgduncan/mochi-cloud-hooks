package mochicloudhooks

import (
	"context"
	"errors"
	"net/http"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/mochi-co/mqtt/v2"
	"github.com/rs/zerolog"
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
			name:        "Success - nil config",
			config:      nil,
			expectError: false,
		},
		{
			name:        "Failure - improper config",
			config:      "",
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
				RoundTripper: mockRT,
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
				RoundTripper: mockRT,
			},
			expectPass: false,
			mocks: func(ctx context.Context) {
				mockRT.EXPECT().RoundTrip(gomock.Any()).Return(nil, errors.New("Oh Crap"))
			},
		},
		{
			name: "Error - Non 2xx",
			config: HTTPAuthHookConfig{
				RoundTripper: mockRT,
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

			success := authHook.OnACLCheck(&mqtt.Client{}, "/topic", false)
			require.Equal(t, tt.expectPass, success)
			// if tt.expectError {
			// 	require.Error(t, err)
			// }

		})
	}
}
