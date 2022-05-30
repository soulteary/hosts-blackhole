# Hosts Blackhole

![](./logo.png)

BEST hosts list processing helper for ad-blocking/blocker programs: AdGuard, AdAway, blocky, grimd...

## Data sources Supported

Currently supported data sources:

- Project: https://github.com/StevenBlack/hosts
  - file: https://raw.githubusercontent.com/StevenBlack/hosts/master/hosts

- Project: https://gitlab.com/quidsup/notrack-blocklists
  - file: https://gitlab.com/quidsup/notrack-blocklists/raw/master/notrack-blocklist.txt
  - file: https://gitlab.com/quidsup/notrack-blocklists/raw/master/notrack-malware.txt

- Project: https://github.com/AdAway/adaway.github.io/
  - file: https://github.com/AdAway/adaway.github.io/blob/master/hosts.txt
  - file: https://adaway.org/hosts.txt

- Project: https://github.com/AdguardTeam/AdGuardSDNSFilter/
  - file: https://adguardteam.github.io/AdGuardSDNSFilter/Filters/filter.txt

## CLI && ENV

CLI parameters:

```bash
  --debug      enable debug mode
  --port int   web port (default 8345)
```

usage: `./hbh --port=8345 --debug=false`

Env variables:

```bash
HBH_DEBUG=true
HBH_PORT=1024
```

usage: `HBH_DEBUG=true HBH_PORT=1024 ./hbh`

## APIs && Pages

API:

- `/data/hosts.txt` Get data file which contains all domains.
- `/ping` Health check
- `/purge` Attempt to reanalyze the data file and generate a cache file

Pages:

- `/` Homepage
- `/list` View available data sources
- `/data` Data source index page

## Screenshot

![](./screenshot/home.png)

![](./screenshot/list.png)
