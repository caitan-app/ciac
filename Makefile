.PHONY: all test clean
.PHONY: linux linux-amd64 darwin darwin-arm64 darwin-amd64

VERSION = $(shell git describe --tags)

clean:
	rm -rf build
all: linux-amd64 darwin-amd64 darwin-arm64
	@echo "all done"

linux-amd64:
	bash build.sh linux amd64 $(VERSION)

darwin-amd64:
	bash build.sh darwin amd64 $(VERSION)

darwin-arm64:
	bash build.sh darwin arm64 $(VERSION)