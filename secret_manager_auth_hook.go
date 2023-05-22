package mochicloudhooks

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"

	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
)

type SecretManagerAuthHook struct {
	usernames []string
	mqtt.HookBase
}

type SecretManagerHookConfig struct {
	Names []string
}

func (h *SecretManagerAuthHook) ID() string {
	return "secret-manager-auth-hook"
}

func (h *SecretManagerAuthHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnACLCheck,
		mqtt.OnConnectAuthenticate,
	}, []byte{b})
}

func (h *SecretManagerAuthHook) Init(config any) error {
	ctx := context.Background()

	if config == nil {
		return errors.New("nil config")
	}

	secretManagerHookConfig, ok := config.(SecretManagerHookConfig)
	if !ok {
		return errors.New("improper config")
	}

	usernames, err := getAdminCredentials(ctx, secretManagerHookConfig.Names)
	if err != nil {
		return err
	}
	h.usernames = usernames

	return nil
}

func (h *SecretManagerAuthHook) OnConnectAuthenticate(cl *mqtt.Client, pk packets.Packet) bool {
	return h.checkAdminCredentials(string(cl.Properties.Username))
}

func (h *SecretManagerAuthHook) OnACLCheck(cl *mqtt.Client, topic string, write bool) bool {
	return h.checkAdminCredentials(string(cl.Properties.Username))
}

func (h *SecretManagerAuthHook) checkAdminCredentials(username string) bool {
	for _, storedUsername := range h.usernames {
		if username == storedUsername {
			return true
		}
	}
	return false
}

func getAdminCredentials(ctx context.Context, names []string) ([]string, error) {
	client, err := secretmanager.NewClient(ctx)
	if err != nil {
		return []string{}, fmt.Errorf("failed to create secretmanager client: %v", err)
	}
	defer client.Close()

	var usernames []string
	for _, name := range names {
		resp, err := client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
			Name: name,
		})
		if err != nil {
			return []string{}, err
		}

		usernames = append(usernames, string(resp.Payload.Data))
	}

	return usernames, nil
}
