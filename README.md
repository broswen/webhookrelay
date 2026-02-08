# WebhookRelay

WebhookRelay is an API that accepts and validates webhook requests, then provisions them to WebhookDispatcher for reliable dispatching.

![diagram](diagram.png)

Sends webhook requests to the [WebhookDispatcher](https://github.com/broswen/webhookdispatcher).

### gRPC API

The gRPC API is defined in `pkg/api/v1/webhookrelay.proto` and provides the following services:

- `GetWebhook` - Retrieve a single webhook by ID
- `ListWebhooks` - List webhooks with pagination
- `CreateWebhook` - Create a new webhook

The gRPC server includes:
- **Reflection Server** - For introspection with tools like `grpcurl`
- **Health Check Server** - For service health monitoring

To regenerate the proto files, run:
```bash
./scripts/gen-proto.sh
```



### Components
- API, a REST API server and gRPC API server used to create and get webhooks.
  - REST API on port 8080 (default)
  - gRPC API on port 8000 (default)
- publisher, an outbox service that publishes new webhooks to a kafka topic for provisioner
- provisioner, a kafka consumer that sends webhook details to the WebhookDispatcher


### TODO
- [ ] create service mocks
- [ ] create repository sql mocks
- [x] create gRPC definition and server