# make generate -C api
.PHONY: generate
generate:
	go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.4.1 \
		--package=v1 --generate=types,chi-server,spec,skip-prune \
		openapi.yaml > openapi.gen.go