BINARY := netgear_cm_exporter

build:
	@go build -o $(BINARY) .

clean:
	@rm -f $(BINARY)
