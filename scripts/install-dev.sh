#! /usr/bin/env zsh
set -e 
trap 'echo "Install Dev Error: Failed at line $LINENO."; exit 1;' ERR

if [[ "$OSTYPE" == "darwin"* ]]; then
    echo "installing dev env on MacOS"
else
    echo "command install-dev only supports MacOS"
    exit
fi   

install_xcode()
{
    xcodebuild -version 2> /dev/null 2>&1
    if [ ! $? -eq 0 ] ; then
        xcode-select --install
    fi
}

install_brew()
{
    which -s brew
    if [ ! $? -eq 0 ] ; then
        echo "installing brew"
        /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
}

is_installed()
{
    command -v "$1" >/dev/null 2>&1
}

install() 
{
    if is_installed $1; then
        echo "$1 already installed"
    else
        echo "installing $"
        brew install $1
    fi
}

add_path()
{
    if [[ ":$PATH:" != *":$1:"* ]]; then
        echo "adding $1 to PATH"
        echo "export PATH=\"$PATH:$1\"" >> ~/.zshrc
    else
        echo "$1 already in PATH"
    fi
}

add_export()
{    
    echo "export $1=$2" >> ~/.zshrc
}

reload_shell()
{
    . ~/.zshrc
}

install_golang()
{
    if is_installed go; then
        echo "go already installed"
    else
        echo "installing go"
        brew install go
        add_export "GOPATH" "$(go env GOPATH)"
        add_export "GOROOT" "$(go env GOROOT)"
        add_export "GOBIN" "$(go env GOROOT)/bin"
        echo 'export PATH="$GOPATH:$PATH"' >> ~/.zshrc
        echo 'export PATH="$GOBIN:$PATH"' >> ~/.zshrc

        install golangci-lint
        reload_shell
    fi
}

install_ginkgo()
{
    go install -mod=mod github.com/onsi/ginkgo/v2/ginkgo@latest
}

ROOT_DIR=$(pwd)

install_xcode
install_brew
install curl
install git
install cmake
install rabbitmq
install_golang
install_ginkgo

cd $ROOT_DIR
