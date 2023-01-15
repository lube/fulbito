all: build

build:
	go build

clean:
	rm -f fulbito

.PHONY: all build clean
