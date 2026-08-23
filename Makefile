.PHONY: build test web clean

build:
	./scripts/build.sh

test:
	./scripts/test.sh

web:
	cd web && npm ci && npm run build

clean:
	go clean

