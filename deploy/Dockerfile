FROM golang:1.27.1 AS build

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY internal/ ./internal/
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/lab-agent ./cmd/lab-agent \
    && CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /out/lab-mcp ./cmd/lab-poisoned-mcp

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /out/lab-agent /lab-agent
COPY --from=build /out/lab-mcp /lab-mcp
