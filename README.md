# WebhookRelay

WebhookRelay is an API that accepts and validates webhook requests, then provisions them to WebhookDispatcher for reliable dispatching.

### Components
- API, a REST API server used to create and get webhooks.
- publisher, an outbox service that publishes new webhooks to a kafka topic for provisioner
- provisioner, a kafka consumer that sends webhook details to the WebhookDispatcher


### TODO
- [ ] wrangler error types up and make everything nice values to compare
    - make http error codes work nicely
- [ ] create helm chart to render k8s manifests
- [ ] create service mocks
- [ ] create repository sql mocks
- [ ] create gRPC definition and server
- [ ] define contract for WebhookDispatcher api
- [ ] add prometheus metrics
- [ ] create openapi spec