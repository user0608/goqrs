VERSION:=$(shell cat version)
image:
	docker build -t ksaucedo/goqrs:$(VERSION) .
push:
	docker push ksaucedo/goqrs:$(VERSION)