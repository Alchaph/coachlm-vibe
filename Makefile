# CoachLM Makefile
# Detects webkit2gtk version and sets build tags accordingly.

WAILS := $(shell which wails 2>/dev/null || echo $(HOME)/go/bin/wails)
TAGS :=

# Use webkit2_41 tag if webkit2gtk-4.1 is available but 4.0 is not
ifeq ($(shell pkg-config --exists webkit2gtk-4.0 2>/dev/null && echo yes),yes)
else ifeq ($(shell pkg-config --exists webkit2gtk-4.1 2>/dev/null && echo yes),yes)
TAGS := webkit2_41
endif

TAG_FLAG := $(if $(TAGS),-tags $(TAGS),)

.PHONY: dev build test clean

dev:
	$(WAILS) dev $(TAG_FLAG)

build:
	$(WAILS) build $(TAG_FLAG)

test:
	go test ./... -count=1

clean:
	rm -rf build/bin
