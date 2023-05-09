linux:
	GOARCH=amd64 GOOS=linux go build -o "bin/pmtaproxy (Linux)"

windows:
	GOARCH=amd64 GOOS=windows go build -o "bin/pmtaproxy.exe (Windows)"

clean:
	rm -r bin/*

all: linux windows