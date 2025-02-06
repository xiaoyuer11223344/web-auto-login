ENV ?= prod
export ENV
export GOPROXY=https://goproxy.cn,direct

build:
	goreleaser build -f ./.goreleaser/build.yml --skip=validate --snapshot --clean
