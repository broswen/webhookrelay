compose:
	docker compose up --build

build:
	docker build . -f Dockerfile.api -t broswen/webhookrelay-api:latest
	docker build . -f Dockerfile.provisioner -t broswen/webhookrelay-provisioner:latest
	docker build . -f Dockerfile.publisher -t broswen/webhookrelay-publisher:latest
