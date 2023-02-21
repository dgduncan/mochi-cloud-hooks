package mochicloudhooks

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

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

type SuperuserCheckPOST struct {
	Username string `json:"username"`
}

type ClientCheckPOST struct {
	ClientID string `json:"clientid"`
	Password string `json:"password"`
	Username string `json:"username"`
}

type ACLCheckPOST struct {
	Username string `json:"username"`
	ClientID string `json:"clientid"`
	Topic    string `json:"topic"`
	ACC      string `json:"acc"`
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
	if config == nil {
		return errors.New("nil config")
	}

	authHookConfig, ok := config.(HTTPAuthHookConfig)
	if !ok {
		return errors.New("improper config")
	}

	fmt.Println(validateConfig(authHookConfig))

	if !validateConfig(authHookConfig) {
		return errors.New("hostname configs failed validation")
	}

	h.httpclient = NewTransport(authHookConfig.RoundTripper)

	h.aclhost = authHookConfig.ACLHost
	h.clientauthhost = authHookConfig.ClientAuthenticationHost
	h.superuserhost = authHookConfig.SuperUserHost
	return nil
}

func (h *HTTPAuthHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	payload := ClientCheckPOST{
		ClientID: cl.ID,
		Password: string(pk.Connect.Password),
		Username: string(pk.Connect.Username),
	}

	resp, err := h.makeRequest(http.MethodPost, h.clientauthhost, payload)
	if err != nil {
		h.Log.Error().Err(err)
		return false
	}

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func (h *HTTPAuthHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	payload := ACLCheckPOST{
		ClientID: cl.ID,
		Username: string(cl.Properties.Username),
		Topic:    topic,
		ACC:      strconv.FormatBool(write),
	}

	resp, err := h.makeRequest(http.MethodPost, h.aclhost, payload)
	if err != nil {
		h.Log.Error().Err(err)
		return false
	}

	return resp.StatusCode >= 200 && resp.StatusCode < 300
}

func (h *HTTPAuthHook) makeRequest(requestType, url string, payload any) (*http.Response, error) {
	var buffer io.Reader
	if payload == nil {
		buffer = http.NoBody
	} else {
		rb, err := json.Marshal(payload)
		if err != nil {
			h.Log.Err(err).Msg("")
			return nil, err
		}
		buffer = bytes.NewBuffer(rb)
	}

	req, err := http.NewRequest(requestType, url, buffer)
	if err != nil {
		h.Log.Error().Err(err)
		return nil, err
	}

	resp, err := h.httpclient.Do(req)
	if err != nil {
		h.Log.Error().Err(err)
		return nil, err
	}

	return resp, nil
}

func validateConfig(config HTTPAuthHookConfig) bool {
	if config.ACLHost == "" || config.ClientAuthenticationHost == "" {
		return false
	}
	return true
}
