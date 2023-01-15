all: build install

build:
	go build

install:
	sudo mv fulbito /usr/local/bin

clean:
	rm -f fulbito

prod-reload:
	sudo systemctl stop fulbito && make && sudo systemctl start fulbito

.PHONY: all build clean prod-reload
