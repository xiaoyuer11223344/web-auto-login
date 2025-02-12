ENV ?= prod
export ENV
export GOPROXY=https://goproxy.cn,direct

mac:
	goreleaser build -f ./.goreleaser/mac.yml --skip=validate --snapshot --clean

linux:
	goreleaser build -f ./.goreleaser/linux.yml --skip=validate --snapshot --clean
