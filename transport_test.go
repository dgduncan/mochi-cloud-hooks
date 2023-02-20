package mochicloudhooks

import (
	"net/http"
	"testing"

	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

// func TestNewRoundTrip(t *testing.T) {

// 	ctrl := gomock.NewController(t)
// 	defer ctrl.Finish()

// 	tests := []struct {
// 		name      string
// 		expectErr bool
// 	}{
// 		{
// 			name:      "Success - Golden Path",
// 			expectErr: false,
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {

// 			req, _ := http.NewRequest("GET", "http://example.com", nil)
// 			nt := NewTransport(http.DefaultTransport)

// 			require.Nil(t, err)
// 		})
// 	}
// }

func TestRoundTrip(t *testing.T) {

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		expectErr bool
	}{
		{
			name:      "Success - Golden Path",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			req, _ := http.NewRequest("GET", "http://example.com", nil)
			nt := new(Transport)
			nt.OriginalTransport = http.DefaultTransport

			_, err := nt.RoundTrip(req)
			if tt.expectErr {
				require.Error(t, err)
				return
			}
			require.Nil(t, err)
		})
	}
}
