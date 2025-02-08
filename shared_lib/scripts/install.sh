#! /usr/bin/env sh

rm go.mod 2> /dev/null

go mod init shared_lib
go mod edit -replace assignment/lib/shared_lib=../shared_lib
go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo@latest 
go mod tidy
go mod download
