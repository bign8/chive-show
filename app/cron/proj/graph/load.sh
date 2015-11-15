#!/bin/sh

if ! which proto >/dev/null; then
	echo "Installing proto and protoc-gen-go"
	go get -u github.com/golang/protobuf/{proto,protoc-gen-go}
	export PATH=$PATH:$GOPATH/bin
else
	echo "Proto and protoc-gen-go already installed"
fi

echo "Generating Protobuff files..."
protoc --go_out=. *.proto
sed -i '' '/RegisterType/d' graph.pb.go
echo "Protobuff files generated."
