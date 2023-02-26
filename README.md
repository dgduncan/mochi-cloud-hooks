[![Go Reference](https://pkg.go.dev/badge/github.com/mochi-co/mqtt.svg)](https://pkg.go.dev/github.com/dgduncan/mochi-cloud-hooks)

# Mochi Cloud Hooks

Mochi Cloud Hooks is a collection of hooks that can be imported and used for Mochi MQTT Broker.
Implementations of certain hooks are inspired by other open source projects

### Table of contents

<!-- MarkdownTOC -->

- [Hooks](#hooks)
    - [Auth](#auth)
        - [HTTP](#http-auth)

<!-- /MarkdownTOC -->

### Hooks

#### Auth

##### HTTP

The HTTP hook is a simple HTTP hook that uses two hooks to authorize the client to connect to the broker and authorizes topic level ACLs.
It works by checking the response code of each endpoint. If an endpoint returns back a non `200` response a `false` is returned to the mochi hook

##### GCP Secret Manager
> :warning: this is currently experimental and should not be used in production

The GCP Secret Manager hook should be utilized as a super admin hook. Secrets stored in Secret Manager will be loaded into memory and compared. If the connecting client matches what is stored in Secret Manager, this user will be a `super user` and will have access to all ACLs. 


