VERSION:=$(shell cat version)
image:
	docker build -t goqrs:$(VERSION) .