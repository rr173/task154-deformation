# 地基形变观测网平差台

服务保存观测网、固定点和待估点，按观测期计算坐标成果；发布后的成果不可变，可与另一观测期进行位移比较。

```bash
GOTOOLCHAIN=local go test ./...
GOTOOLCHAIN=local go vet ./...
CGO_ENABLED=0 GOTOOLCHAIN=local go build ./...
GOTOOLCHAIN=local go run ./cmd/deformation --smoke-test
GOTOOLCHAIN=local go run ./cmd/deformation --addr :8080
```

通过 `./build_benzhi_docker.sh deformation-network linux/amd64` 构建镜像；随后使用 `docker run --rm deformation-network --smoke-test` 执行无外部依赖的容器自检。镜像入口默认执行 `--smoke-test`；服务模式可覆盖参数为 `--addr :8080`。
