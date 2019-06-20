# Paths
# resolve the makefile's path and directory, from https://stackoverflow.com/a/18137056
export MAKE_PATH	?= $(abspath $(lastword $(MAKEFILE_LIST)))
export ROOT_PATH	?= $(dir $(MAKE_PATH))
export TARGET_PATH	?= $(ROOT_PATH)/out

all: configure build

build: docs/example.png out/makegraph

configure:
	mkdir -p out

docs/example.png:
	go run cmd/makegraph/main.go | dot -Tpng > docs/example.png

out/makegraph: configure
	go build -o $@ cmd/makegraph/main.go

.PHONY: clean
clean:
	rm -f docs/example.png
	rm -rf $(TARGET_PATH)