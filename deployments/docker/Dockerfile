FROM golang:1.26 AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/exchange-worker ./cmd/exchange-worker && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/scheduler ./cmd/scheduler && \
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/health-probe ./cmd/health-probe

FROM gcr.io/distroless/base-debian12

WORKDIR /app
COPY --from=build /out/exchange-worker /app/exchange-worker
COPY --from=build /out/scheduler /app/scheduler
COPY --from=build /out/health-probe /app/health-probe
COPY configs /app/configs

ENTRYPOINT ["/app/exchange-worker"]
