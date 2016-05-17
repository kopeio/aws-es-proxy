all: image

code:
	glide install
	go install .

builder-image:
	docker build -f images/builder/Dockerfile -t builder .

build-in-docker: builder-image
	docker run -it -v `pwd`:/src builder /onbuild.sh

image: build-in-docker
	docker build -t kope/aws-es-proxy  -f images/aws-es-proxy/Dockerfile .

push: image
	docker push kope/aws-es-proxy:latest
