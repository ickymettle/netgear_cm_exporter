# Netgear Cable Modem Exporter

Prometheus exporter for Netgear cable modems. This has been developed against a CM600, I suspect it
is likely to work with other modems in the CMxxxx family. If you are able to run this successfully
on another Netgear cable modem model, please see the contributing section below.

## Supported Devices

These Netgear models have been tested and are officially supported:

* Netgear CM600

## Installation

You can build and install the exporter locally by running:

```
go get github.com/ickymettle/netgear_cm_exporter
```

## Usage

```
Usage of netgear_cm_exporter:
  -config.file string
    	Path to configuration file. (optional)
  -modem.address string
    	Cable modem admin administrative ip address and port. (default "192.168.100.1")
  -modem.password string
    	Modem admin password.
  -modem.username string
    	Modem admin username. (default "admin")
  -telemetry.addr string
    	Listen address for metrics endpoint. (default "localhost:9526")
  -telemetry.path string
    	Path to metric exposition endpoint. (default "/metrics")
```

The minimal set of command line flags are the IP address of your cable modem, and the admin password. This
exporter supports a few different means of setting configuration options, you can chose what works best for your environment.

### Configuring via command line flags

```
./netgear_cm_exporter -modem.address 10.0.0.1 -modem.username admin -modem.password foobaz
```

### Configuring via environment variables

Each command line flag can be set in the environment by prefixing the flag with `NETGEAR_CM_EXPORTER` and
providing the command line flag name in uppercase.

eg.

```
export NETGEAR_CM_EXPORTER_MODEM_ADDRESS=10.0.0.1
export NETGEAR_CM_EXPORTER_MODEM_USERNAME=admin
export NETGEAR_CM_EXPORTER_MODEM_PASSWORD=foobaz
```

### Configuring via config file

Lastly if you prefer you can write a config file with each option listed per line in key value pairs delimited by
spaces.

eg. create a file `netgear_cm_exporter.conf` with the following contents:

```
modem.address 10.0.0.1
modem.username admin
modem.password foobaz
```

```
./netgear_exporter -config.file netgear_cm_exporter.conf
```

## Contributing

For the most part I assume the status pages for most of the modems in the CMxxxx family to be
consistent, if you have a Netgear cable modem that isn't properly supported and you're happy to
submit a PR, by all means do so.

If you'd like me to try adding support please do the following:

1. Send me the URL to the modem's status page, this may require some sleuthing in your browser's
   developer console. Typically it's called via JavaScript from the "Cable Connection" link in
   the admin UI. On the CM600 the URI is `/DocsisStatus.asp`.
2. Once you're found the URI to the admin page, send me the saved HTML from that page and I can
   look at what changes are needed to the scraper to support your model.
   
If you do get this running without modification on a CMxxx modem other than the CM600 please submit
a PR to add your model to the list of supported models.
