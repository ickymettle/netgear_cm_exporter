package main

import (
	"encoding/base64"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
	"github.com/gocolly/colly"
	"github.com/peterbourgon/ff"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace = "netgear_cm"
	version   = "0.1.0"
)

type Exporter struct {
	url, authHeaderValue string

	mu sync.Mutex

	// Exporter metrics.
	totalScrapes prometheus.Counter
	scrapeErrors prometheus.Counter

	// Downstream metrics.
	dsChannelSNR               *prometheus.Desc
	dsChannelPower             *prometheus.Desc
	dsChannelCorrectableErrs   *prometheus.Desc
	dsChannelUncorrectableErrs *prometheus.Desc

	// Upstream metrics.
	usChannelPower      *prometheus.Desc
	usChannelSymbolRate *prometheus.Desc
}

// basicAuth returns the base64 encoding of the username and password seperated by
// a colon. Borrowed the net/http package.
func basicAuth(username, password string) string {
	auth := fmt.Sprintf("%s:%s", username, password)
	return base64.StdEncoding.EncodeToString([]byte(auth))
}

func NewExporter(addr, username, password string) *Exporter {
	var (
		dsLabelNames = []string{"channel", "lock_status", "modulation", "channel_id", "frequency"}
		usLabelNames = []string{"channel", "lock_status", "channel_type", "channel_id", "frequency"}
	)

	return &Exporter{
		// Modem access details.
		url:             "http://" + addr + "/DocsisStatus.asp",
		authHeaderValue: "Basic " + basicAuth(username, password),

		// Collection metrics.
		totalScrapes: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "status_scrapes_total",
			Help:      "Total number of scrapes of the modem status page.",
		}),
		scrapeErrors: prometheus.NewCounter(prometheus.CounterOpts{
			Namespace: namespace,
			Name:      "status_scrape_errors_total",
			Help:      "Total number of failed scrapes of the modem status page.",
		}),

		// Downstream metrics.
		dsChannelSNR: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "downstream_channel", "snr_db"),
			"Downstream channel signal to noise ratio in dB.",
			dsLabelNames, nil,
		),
		dsChannelPower: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "downstream_channel", "power_dbmv"),
			"Downstream channel power in dBmV.",
			dsLabelNames, nil,
		),
		dsChannelCorrectableErrs: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "downstream_channel", "correctable_errors_total"),
			"Downstream channel correctable errors.",
			dsLabelNames, nil,
		),
		dsChannelUncorrectableErrs: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "downstream_channel", "uncorrectable_errors_total"),
			"Downstream channel uncorrectable errors.",
			dsLabelNames, nil,
		),

		// Upstream metrics.
		usChannelPower: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "upstream_channel", "power_dbmv"),
			"Upstream channel power in dBmV.",
			usLabelNames, nil,
		),
		usChannelSymbolRate: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "upstream_channel", "power_symbol_rate"),
			"Upstream channel symbol rate per second",
			usLabelNames, nil,
		),
	}
}

func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	// Exporter metrics.
	ch <- e.totalScrapes.Desc()
	ch <- e.scrapeErrors.Desc()
	// Downstream metrics.
	ch <- e.dsChannelSNR
	ch <- e.dsChannelPower
	ch <- e.dsChannelCorrectableErrs
	ch <- e.dsChannelUncorrectableErrs
	// Upstream metrics.
	ch <- e.usChannelPower
	ch <- e.usChannelSymbolRate
}

func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	e.totalScrapes.Inc()

	c := colly.NewCollector()

	// OnRequest callback adds basic auth header.
	c.OnRequest(func(r *colly.Request) {
		r.Headers.Add("Authorization", e.authHeaderValue)
	})

	// OnError callback counts any errors that occur during scraping.
	c.OnError(func(r *colly.Response, err error) {
		e.scrapeErrors.Inc()
	})

	// Callback to parse the tbody block of table with id=dsTable, the downstream table info.
	c.OnHTML(`#dsTable tbody`, func(elem *colly.HTMLElement) {
		elem.DOM.Find("tr").Each(func(i int, row *goquery.Selection) {
			if i == 0 {
				return // no rows were returned
			}
			var (
				channel    string
				lockStatus string
				modulation string
				channelID  string
				freqMHz    string
				snr        float64
				power      float64
				corrErrs   float64
				unCorrErrs float64
			)
			row.Find("td").Each(func(j int, col *goquery.Selection) {
				switch j {
				case 0:
					channel = strings.TrimSpace(col.Text())
				case 1:
					lockStatus = strings.TrimSpace(col.Text())
				case 2:
					modulation = strings.TrimSpace(col.Text())
				case 3:
					channelID = strings.TrimSpace(col.Text())
				case 4:
					{
						var freqHZ float64
						fmt.Sscanf(strings.TrimSpace(col.Text()), "%f Hz", &freqHZ)
						freqMHz = fmt.Sprintf("%0.2f MHz", freqHZ/1e6)
					}
				case 5:
					fmt.Sscanf(strings.TrimSpace(col.Text()), "%f dBmV", &power)
				case 6:
					fmt.Sscanf(strings.TrimSpace(col.Text()), "%f dB", &snr)
				case 7:
					fmt.Sscanf(strings.TrimSpace(col.Text()), "%f", &corrErrs)
				case 8:
					fmt.Sscanf(strings.TrimSpace(col.Text()), "%f", &unCorrErrs)
				}
			})
			labels := []string{channel, lockStatus, modulation, channelID, freqMHz}

			ch <- prometheus.MustNewConstMetric(e.dsChannelSNR, prometheus.GaugeValue, snr, labels...)
			ch <- prometheus.MustNewConstMetric(e.dsChannelPower, prometheus.GaugeValue, power, labels...)
			ch <- prometheus.MustNewConstMetric(e.dsChannelCorrectableErrs, prometheus.CounterValue, corrErrs, labels...)
			ch <- prometheus.MustNewConstMetric(e.dsChannelUncorrectableErrs, prometheus.CounterValue, unCorrErrs, labels...)
		})
	})

	// Callback to parse the tbody block of table with id=usTable, the upstream channel info.
	c.OnHTML(`#usTable tbody`, func(elem *colly.HTMLElement) {
		elem.DOM.Find("tr").Each(func(i int, row *goquery.Selection) {
			if i == 0 {
				return // no rows were returned
			}
			var (
				channel     string
				lockStatus  string
				channelType string
				channelID   string
				symbolRate  float64
				freqMHz     string
				power       float64
			)
			row.Find("td").Each(func(j int, col *goquery.Selection) {
				switch j {
				case 0:
					channel = strings.TrimSpace(col.Text())
				case 1:
					lockStatus = strings.TrimSpace(col.Text())
				case 2:
					channelType = strings.TrimSpace(col.Text())
				case 3:
					channelID = strings.TrimSpace(col.Text())
				case 4:
					{
						var kSymRate float64
						fmt.Sscanf(strings.TrimSpace(col.Text()), "%f Ksym/sec", &kSymRate)
						symbolRate = kSymRate * 1000 // convert to sym/sec
					}
				case 5:
					{
						var freqHZ float64
						fmt.Sscanf(strings.TrimSpace(col.Text()), "%f Hz", &freqHZ)
						freqMHz = fmt.Sprintf("%0.2f MHz", freqHZ/1e6)
					}
				case 6:
					fmt.Sscanf(strings.TrimSpace(col.Text()), "%f dBmV", &power)
				}
			})
			labels := []string{channel, lockStatus, channelType, channelID, freqMHz}

			ch <- prometheus.MustNewConstMetric(e.usChannelPower, prometheus.GaugeValue, power, labels...)
			ch <- prometheus.MustNewConstMetric(e.usChannelSymbolRate, prometheus.GaugeValue, symbolRate, labels...)
		})
	})

	e.mu.Lock()
	c.Visit(e.url)
	e.totalScrapes.Collect(ch)
	e.scrapeErrors.Collect(ch)
	e.mu.Unlock()
}

func main() {
	var (
		fs            = flag.NewFlagSet("netgear_cm_exporter", flag.ExitOnError)
		listenAddress = fs.String("telemetry.addr", "localhost:9526", "Listen address for metrics endpoint.")
		metricsPath   = fs.String("telemetry.path", "/metrics", "Path to metric exposition endpoint.")
		modemAddress  = fs.String("modem.address", "192.168.100.1", "Cable modem admin administrative ip address and port.")
		modemUsername = fs.String("modem.username", "admin", "Modem admin username.")
		modemPassword = fs.String("modem.password", "", "Modem admin password.")
		_             = fs.String("config.file", "", "Path to configuration file. (optional)")
	)

	ff.Parse(fs, os.Args[1:],
		ff.WithConfigFileFlag("config.file"),
		ff.WithConfigFileParser(ff.PlainParser),
		ff.WithEnvVarPrefix("NETGEAR_CM_EXPORTER"),
	)

	exporter := NewExporter(*modemAddress, *modemUsername, *modemPassword)

	prometheus.MustRegister(exporter)

	http.Handle(*metricsPath, prometheus.Handler())
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, *metricsPath, http.StatusMovedPermanently)
	})

	log.Printf("exporter listening on %s", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("failed to start netgear exporter: %s", err)
	}
}
