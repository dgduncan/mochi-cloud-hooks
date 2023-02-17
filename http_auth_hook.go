package mochicloudhooks

import (
	"bytes"
	"errors"
	"net/http"

	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
)

type HTTPAuthHook struct {
	httpclient     *http.Client
	aclhost        string
	clientauthhost string
	superuserhost  string // currently unused
	mqtt.HookBase
}

type HTTPAuthHookConfig struct {
	ACLHost                  string
	SuperUserHost            string
	ClientAuthenticationHost string // currently unused
	RoundTripper             http.RoundTripper
}

func (h *HTTPAuthHook) ID() string {
	return "http-auth-hook"
}

func (h *HTTPAuthHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnACLCheck,
		mqtt.OnConnectAuthenticate,
	}, []byte{b})
}

func (h *HTTPAuthHook) Init(config any) error {
	authHookConfig, ok := config.(HTTPAuthHookConfig)
	if !ok {
		return errors.New("improper config")
	}

	h.httpclient = NewTransport(authHookConfig.RoundTripper)

	h.aclhost = authHookConfig.ACLHost
	h.clientauthhost = authHookConfig.ClientAuthenticationHost
	h.superuserhost = authHookConfig.SuperUserHost
	return nil
}

func (h *HTTPAuthHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	req, err := http.NewRequest(http.MethodGet, h.clientauthhost, http.NoBody)
	if err != nil {
		h.Log.Error().Err(err)
		return false
	}

	resp, err := h.httpclient.Do(req)
	if err != nil {
		h.Log.Error().Err(err)
		return false
	}

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func (h *HTTPAuthHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	req, err := http.NewRequest(http.MethodGet, h.aclhost, http.NoBody)
	if err != nil {
		h.Log.Error().Err(err)
		return false
	}

	resp, err := h.httpclient.Do(req)
	if err != nil {
		h.Log.Error().Err(err)
		return false
	}

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}
