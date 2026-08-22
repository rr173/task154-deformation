FROM docker.m.daocloud.io/library/golang:1.26.3-bookworm AS build

WORKDIR /src
ENV GOPROXY=https://goproxy.cn,direct GOSUMDB=sum.golang.google.cn GOTOOLCHAIN=local
COPY go.mod go.sum ./
RUN go mod download
COPY . .
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -o /out/deformation ./cmd/deformation

FROM docker.m.daocloud.io/library/alpine:3.20
WORKDIR /app
COPY --from=build /out/deformation /app/deformation
ENTRYPOINT ["/app/deformation"]
CMD ["--smoke-test"]
