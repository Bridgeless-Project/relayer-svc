protogen:
	cd proto && \
	buf generate deposit --template=./templates/deposit.yaml --config=buf.yaml && \
	buf generate api --template=./templates/api.yaml --config=buf.yaml


install:
	export CGO_ENABLED=1
	rm -f $(GOPATH)/bin/relayer-svc
	go build -o $(GOPATH)/bin

run: install
	relayer-svc service run