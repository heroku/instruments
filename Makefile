default: vet test

test:
	go test ./...

test-race:
	go test ./... -race

vet:
	go vet ./...

bench:
	go test ./... -test.run=NONE -test.bench=. -test.benchmem
