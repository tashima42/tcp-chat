FROM golang:1.24 AS build

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -ldflags="-s -w" -o tcp-chat .

FROM scratch AS final

WORKDIR /

COPY --from=build /app/tcp-chat /

USER nonroot:nonroot

ENTRYPOINT ["/tcp-chat"]
