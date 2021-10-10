FROM docker.io/library/golang:1.16 AS builder
WORKDIR /src
COPY . /src
RUN go build

FROM docker.io/library/debian:stable-slim
COPY --from=builder /src/lsmsd /
WORKDIR /
RUN ls -lah
ENTRYPOINT [ "/lsmsd" ]