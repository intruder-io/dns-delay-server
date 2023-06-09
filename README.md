This is a simple DNS server which can be used to delay A or AAAA responses. This is useful for quickly exploiting DNS rebinding in Safari, as described in the accompanying [research blog](https://intruder.io/research/tricks-for-split-second-dns-rebinding).

# Installation
Ensure that you have golang properly installed and setup, then install this tool with:
```
go get github.com/intruderio/dns-delay-server@latest
```

# Usage
To use this server you will need a domain, and you will need to setup a corresponding NS record for that domain to point to this server. For example, if you own `example.com` and wanted this DNS server to handle all queries for `*.r.example.com`, you would need to:

1. Run this server on a VPS
2. Create a hostname for that VPS, such as `vps.example.com`, with corresponding `A` and `AAAA` records in your DNS provider
3. Create an NS record for `*.r.example.com` with a valu of `vps.example.com`

To have this server respond to all `A` queries with `1.2.3.4` and all `AAAA` queries with `::1`, and delay all `AAAA` responses by 200ms, you would run the following as root:
```
dns-delay-server -p 53 -a 1.2.3.4 -6 ::1 -D 200ms
```

It is common to already have a DNS server listening on a local adapter. In this case, you can avoid clashes by specifying the IP address to listen on with `-l`.

The full help text can be seen with `dns-delay-server -h`.

## CNAMEs
The server will respond to `cname.<your-domain>` with a CNAME record to `<your-domain>`. If you want these responses to return different IP addresses to `A`/`AAAA` queries, you can use the `-c` and `-C` flags to specify the IP addresses to include in CNAME responses, or prevent them being included at all.
