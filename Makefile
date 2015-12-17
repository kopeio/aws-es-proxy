all: code image

code:
	go install .

image: code
	mkdir -p .build/
	cp ${GOPATH}/bin/aws-es-proxy .build/
	docker build -t kope/aws-es-proxy .

push: image
	docker push kope/aws-es-proxy:latest
