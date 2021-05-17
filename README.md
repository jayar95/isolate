# isolate

## About

isolate is a declarative application server / reverse proxy that includes
GSL (generic language support) out of the box. Similar to containers,
isolate will create network namespaces for each application defined in
an [isolate config file](example-config.json) which will isolate a given program
to it's own network namespace with a dedicated local ip address + netlinks.

This was a sunday project so it's far from production-ready.

## Configuration Example

```json
{
    "applications": [
        {
            "name": "deliveries",
            "executable": "/apps/express/deliveries",
            "port": "80",
            "route": "/deliveries",
            "scheme": "http"
        },
        {
            "name": "dashboard",
            "executable": "/apps/express/run",
            "port": "80",
            "route": "/*dashboard",
            "scheme": "http"
        }
    ]
}
```

You can also view an example config at [isolate config file](example-config.json)

## Requirements

- go 1.16
- `isolate` requires elevate privileges to run.
- A linux operating system (Darkwin/Windows not tested)

## Build

```bash
make build
```

The binary `multi-dev` should be available in the working directory

## Usage

```bash
CONFIG="/var/apps/isolate-server/isolate-config.json"
./isolate $CONFIG
```

If `$CONFIG` is not specified, isolate will assume a file `isolate-config.json`
is present in the program's directory and try to use that.

## Future enhancements
- Add NAT to provide external internet connectivity for all isolated applications / network namespaces !important
- Unit tests
- TLS support
- Add port config to `isolate` to customize what port the server listens on