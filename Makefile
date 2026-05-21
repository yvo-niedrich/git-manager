BINARY  := gitmg
CMD     := ./cmd/gitmg
DIST    := dist

PLATFORMS := \
	darwin/amd64 \
	darwin/arm64 \
	linux/amd64 \
	linux/arm64 \
	windows/amd64

NAME    ?= git-mg
PREFIX  ?= /usr/local/bin

.PHONY: build test run install clean

build:
	@mkdir -p $(DIST)
	@$(foreach platform,$(PLATFORMS), \
		$(eval OS   := $(word 1,$(subst /, ,$(platform)))) \
		$(eval ARCH := $(word 2,$(subst /, ,$(platform)))) \
		$(eval EXT  := $(if $(filter windows,$(OS)),.exe,)) \
		$(eval OUT  := $(DIST)/$(BINARY)-$(OS)-$(ARCH)$(EXT)) \
		echo "→ $(OUT)" && \
		GOOS=$(OS) GOARCH=$(ARCH) go build -o $(OUT) $(CMD) && \
	) true

install:
	go build -o $(PREFIX)/$(NAME) $(CMD)
	@echo "→ installed $(PREFIX)/$(NAME)"

test:
	go test ./...

run:
	go run $(CMD)

clean:
	rm -rf $(DIST)
