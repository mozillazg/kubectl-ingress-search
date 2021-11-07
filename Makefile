
CGO_ENABLED ?= 0

.PHONY: build
build:
	CGO_ENABLED=$(CGO_ENABLED) go build -a -o kubectl-ingress-search cmd/kubectl-ingress-search/main.go
