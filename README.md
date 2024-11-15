# Proxy

Collection of proxy implementations.

# Install

Build, for example, SOCKS5:

```
$ make build-socks5
```

Install. Most likely you will need `sudo`, `doas` or somethink like that before the command:

```
$ make install-socks5
```

# Usage

```
$ socks5 -help
```

# Implementations

- Simple implementation of [SOCKS5](https://www.ietf.org/rfc/rfc1928.txt) proxy: establishing a TCP connection without authentication and IPv6.
