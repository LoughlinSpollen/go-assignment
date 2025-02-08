#! /usr/bin/env sh

install()
{
    local CURRENT_DIR=$(pwd)
    local PROJECT_ROOT=$(git rev-parse --show-toplevel)
    cd $PROJECT_ROOT
    if [ ! -f "$PROJECT_ROOT/shared_lib/go.mod" ]; then
        echo "shared_lib module not found. Installing shared_lib module"
        cd $PROJECT_ROOT/shared_lib
        sh scripts/install.sh
    fi

    cd $PROJECT_ROOT/service
    rm go.mod 2>/dev/null
    
    go mod init assignment_service
    go mod edit -replace assignment/lib/shared_lib=../shared_lib	
    go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo@latest
    go mod tidy
    go mod download

    cd $CURRENT_DIR
}

install
