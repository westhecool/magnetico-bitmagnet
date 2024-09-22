# Magnetico-Bitmagnet
Automatically discover torrents using magnetico and immediately import them into bitmagnet.

## Example usage:
With docker compose:
```yaml
services:
  magnetico:
    image: ghcr.io/westhecool/magnetico-bitmagnet:latest
    # the import url will be the address that you use to access bitmagnet's ui plus "/import" for example: "http://localhost:3333/import" (note: your localhost is not going to be available in Docker containers)
    command: --import-url=IMPORT_URL
    restart: unless-stopped
```

## Credits
Thanks to [tgragnato/magnetico](https://github.com/tgragnato/magnetico) for almost all the code.