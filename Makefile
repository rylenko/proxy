.POSIX:

# TODO: linter

include config.mk

all: build

build: build-socks5

build-socks5:
	go build -o $(SOCKS5_NAME) $(SOCKS5_CMD)

clean:
	rm -rf $(SOCKS5_NAME)

install: install-socks5

install-socks5: $(SOCKS5_NAME)
	mkdir -p $(PREFIX)/bin
	cp $(SOCKS5_NAME) $(PREFIX)/bin/$(SOCKS5_NAME)
	chmod 755 $(PREFIX)/bin/$(SOCKS5_NAME)

uninstall: uninstall-socks5

uninstall-socks5:
	rm $(PREFIX)/bin/$(SOCKS5_NAME)

.PHONY: all build build-socks5 clean install install-socks5 uninstall \
	uninstall-socks5
