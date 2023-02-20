package mochicloudhooks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestID(t *testing.T) {
	authHook := new(HTTPAuthHook)

	require.Equal(t, "http-auth-hook", authHook.ID())
}
