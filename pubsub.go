package mochicloudhooks

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/mochi-co/mqtt/v2"
	"github.com/mochi-co/mqtt/v2/packets"
)

type PubsubMessagingHook struct {
	connectTopic   *pubsub.Topic
	publishTopic   *pubsub.Topic
	subscripeTopic *pubsub.Topic
	willTopic      *pubsub.Topic
	disallowlist   []string
	mqtt.HookBase
}

type PubsubMessagingHookConfig struct {
	ProjectID          string
	PublishTopicName   string
	SubscribeTopicName string
	ConnectTopicName   string
	WillTopicName      string
	DisallowList       []string
}

type PublishMessage struct {
	ClientID  string    `json:"client_id"`
	Topic     string    `json:"topic"`
	Payload   []byte    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

type ConnectMessage struct {
	ClientID  string    `json:"client_id"`
	Username  string    `json:"username"`
	Timestamp time.Time `json:"timestamp"`
	Connected bool      `json:"connected"`
}

type SubscribeMessage struct {
	ClientID   string    `json:"client_id"`
	Username   string    `json:"username"`
	Topic      string    `json:"topic"`
	Subscribed bool      `json:"subscribed"`
	Timestamp  time.Time `json:"timestamp"`
}

type OnWillSentMessage struct {
	ClientID  string    `json:"client_id"`
	Topic     string    `json:"topic"`
	Payload   []byte    `json:"payload"`
	Timestamp time.Time `json:"timestamp"`
}

func (pmh *PubsubMessagingHook) ID() string {
	return "pubsub-messaging-hook"
}

func (pmh *PubsubMessagingHook) Provides(b byte) bool {
	return bytes.Contains([]byte{
		mqtt.OnConnect,
		mqtt.OnDisconnect,
		mqtt.OnPublished,
		mqtt.OnSubscribed,
		mqtt.OnUnsubscribed,
		mqtt.OnWillSent,
	}, []byte{b})
}

func (pmh *PubsubMessagingHook) Init(config any) error {
	ctx := context.Background()

	if config == nil {
		return errors.New("nil config")
	}

	pubsubMessagingHookConfig, ok := config.(PubsubMessagingHookConfig)
	if !ok {
		return errors.New("improper config")
	}

	// Create and configure pubsub client
	pc, err := pubsub.NewClient(ctx, pubsubMessagingHookConfig.ProjectID)
	if err != nil {
		pmh.Log.Err(err).Msg("failed to create pubsub client")
		return errors.New("failed to create pubsub client")
	}

	if pubsubMessagingHookConfig.PublishTopicName != "" {
		pubslishtopic := pc.Topic(pubsubMessagingHookConfig.PublishTopicName)
		pubslishtopic.PublishSettings = pubsub.PublishSettings{
			DelayThreshold: 1 * time.Second,
			CountThreshold: 10,
		}
		pmh.publishTopic = pubslishtopic
	}

	if pubsubMessagingHookConfig.SubscribeTopicName != "" {
		subscribetopic := pc.Topic(pubsubMessagingHookConfig.SubscribeTopicName)
		subscribetopic.PublishSettings = pubsub.PublishSettings{
			DelayThreshold: 1 * time.Second,
			CountThreshold: 10,
		}
		pmh.subscripeTopic = subscribetopic
	}

	if pubsubMessagingHookConfig.ConnectTopicName != "" {
		connecttopic := pc.Topic(pubsubMessagingHookConfig.ConnectTopicName)
		connecttopic.PublishSettings = pubsub.PublishSettings{
			DelayThreshold: 1 * time.Second,
			CountThreshold: 10,
		}
		pmh.connectTopic = connecttopic
	}

	if pubsubMessagingHookConfig.WillTopicName != "" {
		willTopic := pc.Topic(pubsubMessagingHookConfig.WillTopicName)
		willTopic.PublishSettings = pubsub.PublishSettings{
			DelayThreshold: 1 * time.Second,
			CountThreshold: 10,
		}
		pmh.willTopic = willTopic
	}

	pmh.disallowlist = pubsubMessagingHookConfig.DisallowList

	return nil
}

func (pmh *PubsubMessagingHook) OnUnsubscribed(cl *mqtt.Client, pk packets.Packet) {
	if pmh.subscripeTopic == nil {
		return
	}

	if !pmh.checkAllowed(string(cl.Properties.Username)) {
		return
	}

	if err := publish(pmh.subscripeTopic, SubscribeMessage{
		ClientID:   cl.ID,
		Username:   string(cl.Properties.Username),
		Timestamp:  time.Now(),
		Subscribed: false,
		Topic:      pk.TopicName,
	}); err != nil {
		pmh.Log.Err(err).Msg("")
	}
}

func (pmh *PubsubMessagingHook) OnSubscribed(cl *mqtt.Client, pk packets.Packet, reasonCodes []byte) {
	if pmh.subscripeTopic == nil {
		return
	}

	if !pmh.checkAllowed(string(cl.Properties.Username)) {
		return
	}

	if err := publish(pmh.subscripeTopic, SubscribeMessage{
		ClientID:   cl.ID,
		Username:   string(cl.Properties.Username),
		Timestamp:  time.Now(),
		Subscribed: true,
		Topic:      pk.TopicName,
	}); err != nil {
		pmh.Log.Err(err).Msg("")
	}
}

func (pmh *PubsubMessagingHook) OnConnect(cl *mqtt.Client, pk packets.Packet) {
	if pmh.connectTopic == nil {
		return
	}

	if !pmh.checkAllowed(string(cl.Properties.Username)) {
		return
	}

	if err := publish(pmh.connectTopic, ConnectMessage{
		ClientID:  cl.ID,
		Username:  string(cl.Properties.Username),
		Timestamp: time.Now(),
		Connected: true,
	}); err != nil {
		pmh.Log.Err(err).Msg("")
	}
}

func (pmh *PubsubMessagingHook) OnDisconnect(cl *mqtt.Client, connect_err error, expire bool) {
	if pmh.connectTopic == nil {
		return
	}

	if !pmh.checkAllowed(string(cl.Properties.Username)) {
		return
	}

	if err := publish(pmh.connectTopic, ConnectMessage{
		ClientID:  cl.ID,
		Username:  string(cl.Properties.Username),
		Timestamp: time.Now(),
		Connected: false,
	}); err != nil {
		pmh.Log.Err(err).Msg("")
	}
}

func (pmh *PubsubMessagingHook) OnPublished(cl *mqtt.Client, pk packets.Packet) {
	if pmh.publishTopic == nil {
		return
	}

	if !pmh.checkAllowed(string(cl.Properties.Username)) {
		return
	}

	if err := publish(pmh.publishTopic, PublishMessage{
		ClientID:  cl.ID,
		Topic:     pk.TopicName,
		Payload:   pk.Payload,
		Timestamp: time.Now(),
	}); err != nil {
		pmh.Log.Err(err).Msg("")
	}
}

func (pmh *PubsubMessagingHook) OnWillSent(cl *mqtt.Client, pk packets.Packet) {
	if pmh.willTopic == nil {
		return
	}

	if !pmh.checkAllowed(string(cl.Properties.Username)) {
		return
	}

	if err := publish(pmh.publishTopic, OnWillSentMessage{
		ClientID:  cl.ID,
		Topic:     pk.TopicName,
		Payload:   pk.Payload,
		Timestamp: time.Now(),
	}); err != nil {
		pmh.Log.Err(err).Msg("")
	}
}

func (pmh *PubsubMessagingHook) checkAllowed(username string) bool {
	for _, disallowedUsername := range pmh.disallowlist {
		if username == disallowedUsername {
			return false
		}
	}
	return true
}

func publish(topic *pubsub.Topic, data any) error {
	ctx := context.Background()
	b, _ := json.Marshal(data)

	// TODO : add options to store response for later
	topic.Publish(ctx, &pubsub.Message{
		Data: b,
	})

	return nil
}
