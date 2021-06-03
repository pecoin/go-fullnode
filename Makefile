# This Makefile is meant to be used by people that do not usually work
# with Go source code. If you know what GOPATH is then you probably
# don't need to bother with make.

.PHONY: fullnode android ios fullnode-cross evm all test clean
.PHONY: fullnode-linux fullnode-linux-386 fullnode-linux-amd64 fullnode-linux-mips64 fullnode-linux-mips64le
.PHONY: fullnode-linux-arm fullnode-linux-arm-5 fullnode-linux-arm-6 fullnode-linux-arm-7 fullnode-linux-arm64
.PHONY: fullnode-darwin fullnode-darwin-386 fullnode-darwin-amd64
.PHONY: fullnode-windows fullnode-windows-386 fullnode-windows-amd64

GOBIN = ./build/bin
GO ?= latest
GORUN = env GO111MODULE=on go run

fullnode:
	$(GORUN) build/ci.go install ./cmd/fullnode
	@echo "Done building."
	@echo "Run \"$(GOBIN)/fullnode\" to launch fullnode."

all:
	$(GORUN) build/ci.go install

android:
	$(GORUN) build/ci.go aar --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/fullnode.aar\" to use the library."
	@echo "Import \"$(GOBIN)/fullnode-sources.jar\" to add javadocs"
	@echo "For more info see https://stackoverflow.com/questions/20994336/android-studio-how-to-attach-javadoc"
	
ios:
	$(GORUN) build/ci.go xcode --local
	@echo "Done building."
	@echo "Import \"$(GOBIN)/Geth.framework\" to use the library."

test: all
	$(GORUN) build/ci.go test

lint: ## Run linters.
	$(GORUN) build/ci.go lint

clean:
	env GO111MODULE=on go clean -cache
	rm -fr build/_workspace/pkg/ $(GOBIN)/*

# The devtools target installs tools required for 'go generate'.
# You need to put $GOBIN (or $GOPATH/bin) in your PATH to use 'go generate'.

devtools:
	env GOBIN= go get -u golang.org/x/tools/cmd/stringer
	env GOBIN= go get -u github.com/kevinburke/go-bindata/go-bindata
	env GOBIN= go get -u github.com/fjl/gencodec
	env GOBIN= go get -u github.com/golang/protobuf/protoc-gen-go
	env GOBIN= go install ./cmd/abigen
	@type "npm" 2> /dev/null || echo 'Please install node.js and npm'
	@type "solc" 2> /dev/null || echo 'Please install solc'
	@type "protoc" 2> /dev/null || echo 'Please install protoc'

# Cross Compilation Targets (xgo)

fullnode-cross: fullnode-linux fullnode-darwin fullnode-windows fullnode-android fullnode-ios
	@echo "Full cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-*

fullnode-linux: fullnode-linux-386 fullnode-linux-amd64 fullnode-linux-arm fullnode-linux-mips64 fullnode-linux-mips64le
	@echo "Linux cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-linux-*

fullnode-linux-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/386 -v ./cmd/fullnode
	@echo "Linux 386 cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-linux-* | grep 386

fullnode-linux-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/amd64 -v ./cmd/fullnode
	@echo "Linux amd64 cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-linux-* | grep amd64

fullnode-linux-arm: fullnode-linux-arm-5 fullnode-linux-arm-6 fullnode-linux-arm-7 fullnode-linux-arm64
	@echo "Linux ARM cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-linux-* | grep arm

fullnode-linux-arm-5:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-5 -v ./cmd/fullnode
	@echo "Linux ARMv5 cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-linux-* | grep arm-5

fullnode-linux-arm-6:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-6 -v ./cmd/fullnode
	@echo "Linux ARMv6 cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-linux-* | grep arm-6

fullnode-linux-arm-7:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm-7 -v ./cmd/fullnode
	@echo "Linux ARMv7 cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-linux-* | grep arm-7

fullnode-linux-arm64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/arm64 -v ./cmd/fullnode
	@echo "Linux ARM64 cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-linux-* | grep arm64

fullnode-linux-mips:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips --ldflags '-extldflags "-static"' -v ./cmd/fullnode
	@echo "Linux MIPS cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-linux-* | grep mips

fullnode-linux-mipsle:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mipsle --ldflags '-extldflags "-static"' -v ./cmd/fullnode
	@echo "Linux MIPSle cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-linux-* | grep mipsle

fullnode-linux-mips64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64 --ldflags '-extldflags "-static"' -v ./cmd/fullnode
	@echo "Linux MIPS64 cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-linux-* | grep mips64

fullnode-linux-mips64le:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=linux/mips64le --ldflags '-extldflags "-static"' -v ./cmd/fullnode
	@echo "Linux MIPS64le cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-linux-* | grep mips64le

fullnode-darwin: fullnode-darwin-386 fullnode-darwin-amd64
	@echo "Darwin cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-darwin-*

fullnode-darwin-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/386 -v ./cmd/fullnode
	@echo "Darwin 386 cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-darwin-* | grep 386

fullnode-darwin-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=darwin/amd64 -v ./cmd/fullnode
	@echo "Darwin amd64 cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-darwin-* | grep amd64

fullnode-windows: fullnode-windows-386 fullnode-windows-amd64
	@echo "Windows cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-windows-*

fullnode-windows-386:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/386 -v ./cmd/fullnode
	@echo "Windows 386 cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-windows-* | grep 386

fullnode-windows-amd64:
	$(GORUN) build/ci.go xgo -- --go=$(GO) --targets=windows/amd64 -v ./cmd/fullnode
	@echo "Windows amd64 cross compilation done:"
	@ls -ld $(GOBIN)/fullnode-windows-* | grep amd64
