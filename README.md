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
Usage of ./netgear_cm_exporter:
  -config.file string
    	Path to configuration file. (default "netgear_cm_exporter.yml")
  -version
    	Print version information.
```

An example configuration file is provided in `netgear_cm_exporter.yml` showing all the possible
configuration options. The values in the example are the defaults, the bare minimum configuration
is the administrative password to your modem:

```
modem:
  password: <your password here>
```

## Grafana Dashboard

A sample grafana dashboard can be found in the `grafana/` directory. You can import `netgear_cable_modem.json` into 
your Grafana instance to get up and running with a quick dashboard.

![Grafana Dashboard Screenshot](/grafana/dashboard_screenshot.png)

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
