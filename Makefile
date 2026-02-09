compose:
	docker compose up --build

build:
	docker build . -f Dockerfile.api -t broswen/webhookrelay-api:latest
	docker build . -f Dockerfile.publisher -t broswen/webhookrelay-publisher:latest
	docker build . -f Dockerfile.provisioner -t broswen/webhookrelay-provisioner:latest

publish: build
	docker push broswen/webhookrelay-api:latest
	docker push broswen/webhookrelay-publisher:latest
	docker push broswen/webhookrelay-provisioner:latest

helm-template:
	helm template webhookrelay k8s/webhookrelay > k8s/deploy/webhookrelay.yaml

gen-proto:
	./scripts/gen-proto.sh

test: helm-template
	go test ./...
	kubeconform -summary -strict ./k8s/deploy/*.yaml