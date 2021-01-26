.PHONY: all clean debug

all:
	go build -o bin/gatecount-api ./cmd/gatecount-api

clean:
	rm -f bin/*

# Requires the delve debugger
debug:
	go build -gcflags '-N -l' -o bin/gatecount-api-debug ./cmd/gatecount-api
	dlv exec ./bin/gatecount-api-debug
