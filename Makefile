all: build install

build:
	go build

install:
	sudo mv fulbito /usr/local/bin

clean:
	rm -f fulbito

.PHONY: all build clean
