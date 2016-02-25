default: test

test:
	go test ./... -race

vet:
	go vet ./...

bench:
	go test ./... -test.run=NONE -test.bench=. -test.benchmem
