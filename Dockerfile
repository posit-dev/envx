FROM golang:bookworm AS golang

ARG TARGETOS=linux
ARG TARGETARCH=amd64

WORKDIR /build
ENV CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH}
COPY go.mod ./
RUN go mod download
COPY . .
RUN go build -tags netgo -a -o envx ./cmd/envx/main.go

FROM scratch

COPY --from=golang /build/envx /bin/envx
CMD ["envx", "--help"]
