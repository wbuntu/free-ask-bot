PKG = gitbub.com/wbuntu/gin-template
IMG ?= docker.io/wbuntu/gin-template
VERSION ?= $(shell date "+v%Y%m%d")
PLATFORM ?= linux/amd64,linux/arm64
COMMIT = $(shell git log --format="%h" -n 1|tr -d '\n')
TIMESTAMP = $(shell date -u "+%Y-%m-%dT%H:%M:%SZ")

# 生成swagger
.PHONY: swagger
swagger:
	@swag fmt && swag init

# 本地构建
.PHONY: build
build:
	@echo "Compiling source"
	@mkdir -p build
	@go build -ldflags "-s -w -X $(PKG)/cmd.version=$(VERSION)-$(COMMIT)-$(TIMESTAMP)" -o build/gin-template main.go

# 本地运行
.PHONY: run
run: build
	@echo "Running gin-template"
	@./build/gin-template

# 安装开发组件 
.PHONY: dev-requirements
dev-requirements:
	@echo "Installing development requirements"
	@go install golang.org/x/tools/cmd/stringer@v0.4.0
# 	swag v1.8.8版本异常，当前必须使用v1.8.7
	@go install github.com/swaggo/swag/cmd/swag@v1.8.7
#	ginkgo版本与依赖版本保持一致
	@go install github.com/onsi/ginkgo/v2/ginkgo@v2.7.0

# 生成枚举类型的String方法
.PHONY: enums
enums:
	@echo "Generating codes for enums"
	@stringer -output internal/storage/enum_string.go -type=TaskStatus internal/storage/enum.go 

# 单元测试
.PHONY: test
test:
	@echo "Running unit test"
	@go test -count=1 -v $(shell go list ./... | grep -Ev "e2e|vendor")

.PHONY: release-test
release-test:
	@mkdir -p report
	@go test -count=1 -v -cover -coverprofile cover.out $(shell go list ./... | grep -Ev "e2e|vendor") | go-junit-report > report/test.xml  
	@go tool cover -html=cover.out -o report/index.html

# 端到端测试
.PHONY: e2e-build
e2e-build:
	@echo "Building e2e test"
	@ginkgo build -ldflags "-s -w" ./e2e

.PHONY: e2e-run
e2e-run: e2e-build
	@echo "Running e2e test"
	@@./e2e/e2e.test --ginkgo.v --ginkgo.timeout=2h

# 使用Docker构建本机架构镜像
.PHONY: image
image: 
	@echo "Building image $(IMG):$(VERSION)"
	@docker buildx build -t $(IMG):$(VERSION) --load .

# 推送镜像
.PHONY: push
push: image
	@echo "Pushing image $(IMG):$(VERSION)"
	@docker push $(IMG):$(VERSION) 

# 使用Docker构建arm64镜像
.PHONY: image-arm64
image-arm64: 
	@echo "Building image $(IMG):$(VERSION)"
	@docker buildx build --platform linux/arm64 -t $(IMG):$(VERSION) --load .

# 推送镜像
.PHONY: push-arm64
push-arm64: image-arm64
	@echo "Pushing image $(IMG):$(VERSION)"
	@docker push $(IMG):$(VERSION) 

# 使用Docker构建amd64镜像
.PHONY: image-amd64
image-amd64: 
	@echo "Building image $(IMG):$(VERSION)"
	@docker buildx build --platform linux/amd64 -t $(IMG):$(VERSION) --load .

# 推送镜像
.PHONY: push-amd64
push-amd64: image-amd64
	@echo "Pushing image $(IMG):$(VERSION)"
	@docker push $(IMG):$(VERSION) 

# 使用Docker构建双架构镜像
.PHONY: release-image
release-image:
	@echo "Building image $(IMG):$(VERSION) with platform $(PLATFORM)"
	@docker buildx build --platform $(PLATFORM) -t $(IMG):$(VERSION) .

# 推送镜像
.PHONY: release-push
release-push:
	@echo "Building image $(IMG):$(VERSION) with platform $(PLATFORM)"
	@docker buildx build --platform $(PLATFORM) -t $(IMG):$(VERSION) --push .