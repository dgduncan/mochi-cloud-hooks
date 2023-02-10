package mochicloudhooks

import (
	"bytes"

	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
)

type HTTPAuthHook struct {
	mqtt.HookBase
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
	return nil
}

func (h *HTTPAuthHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {

	return true
}

func (h *HTTPAuthHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	return true
}
