all: build

build:
	go build

clean:
	rm -f app

.PHONY: all build clean
