FROM golang:1.24.5-alpine AS build

WORKDIR /build

COPY go.mod go.mod ./

RUN go mod download

COPY . ./

RUN go build -o ./riotRateLimiting


FROM gcr.io/distroless/base

WORKDIR /rateLimiting

COPY --from=build /build/riotRateLimiting ./riotRateLimiting

CMD [ "/rateLimiting/riotRateLimiting" ]