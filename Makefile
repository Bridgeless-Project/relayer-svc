protogen:
	cd proto && \
	buf generate deposit --template=./templates/deposit.yaml --config=buf.yaml && \
	buf generate api --template=./templates/api.yaml --config=buf.yaml