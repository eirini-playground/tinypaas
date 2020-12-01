CLI_PATH = $(CURDIR)/cli
all: build

build:
	cd $(CLI_PATH) && go build -o ../dist/cli
