default: build

BINARY_NAME=terraform-provider-mihari
HOSTNAME=registry.terraform.io
NAMESPACE=mihari-io
NAME=mihari
VERSION=0.1.0
OS_ARCH=$(shell go env GOOS)_$(shell go env GOARCH)

build:
	go build -o $(BINARY_NAME)

install: build
	mkdir -p ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)
	cp $(BINARY_NAME) ~/.terraform.d/plugins/$(HOSTNAME)/$(NAMESPACE)/$(NAME)/$(VERSION)/$(OS_ARCH)/

test:
	go test ./internal/... -v -count=1 -timeout 120s

testacc:
	TF_ACC=1 go test ./internal/... -v -count=1 -timeout 600s

fmt:
	gofmt -s -w .

tidy:
	go mod tidy

docs:
	go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate -provider-name mihari

clean:
	rm -f $(BINARY_NAME)

.PHONY: build install test testacc fmt tidy clean docs
