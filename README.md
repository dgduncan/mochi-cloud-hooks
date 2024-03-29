[![Go Reference](https://pkg.go.dev/badge/github.com/mochi-co/mqtt.svg)](https://pkg.go.dev/github.com/dgduncan/mochi-cloud-hooks)

# Mochi Cloud Hooks

Mochi Cloud Hooks is a collection of hooks that can be imported and used for Mochi MQTT Broker.
Implementations of certain hooks are inspired by other open source projects

### Table of contents

<!-- MarkdownTOC -->

- [Hooks](#hooks)
    - [Auth](#auth)
        - [HTTP](#http-auth)
        - [GCP Secret Manager](#gcp-secret-manager)
    - [Messaging](#messaging)
        - [Pub/Sub](#pubsub)
    

<!-- /MarkdownTOC -->

### Hooks

#### Auth

##### HTTP

The HTTP hook is a simple HTTP hook that uses two hooks to authorize the client to connect to the broker and authorizes topic level ACLs.
It works by checking the response code of each endpoint. If an endpoint returns back a non `200` response a `false` is returned from the hook

##### GCP Secret Manager
> :warning: this is currently experimental and should not be used in production. The functionality is purly for testing and will be changed in the future

The GCP Secret Manager hook should be utilized as a super admin hook. Secrets stored in Secret Manager will be loaded into memory and compared at runtime. If the connecting client's username matches what is stored in Secret Manager, this user will be a `super user` and will have access to all ACLs. 

#### Messaging

##### Pub/Sub

The Pub/Sub hook uses GCP Pub/Sub to publish messages to topics for subscribing, publishing, and connecting. An optional disallow list can be passed in that will check if the username responsible for the event should be allowed to publish to the topic. This is done to prevent overloading from admin clients that may be responsible for a large amount of messages, connections, or subscriptions.


