FROM golang:1.25 AS build
WORKDIR /src
COPY . .
RUN go mod download
RUN go build -o /out/wow ./cmd/wow

FROM ghcr.io/charmbracelet/vhs:latest
WORKDIR /repo
COPY --from=build /out/wow /usr/local/bin/wow
ENV WOW_HOME=/tmp/wow-home
ENTRYPOINT ["vhs"]


