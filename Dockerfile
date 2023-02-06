FROM golang:1.20-buster AS build

WORKDIR /app

COPY go.mod ./

COPY go.sum ./

RUN go mod download

COPY *.go ./

COPY internal ./internal

RUN go build -o /dependency-track-exporter

FROM gcr.io/distroless/static:nonroot

WORKDIR /

COPY --from=build /dependency-track-exporter /dependency-track-exporter

EXPOSE 9916

USER nonroot:nonroot

ENTRYPOINT ["/dependency-track-exporter"]