FROM golang:1.24.5-alpine AS build

WORKDIR /build

COPY go.mod go.mod ./

RUN go mod download

COPY . ./

RUN go build -o ./initializePlayers


FROM gcr.io/distroless/base

WORKDIR /initialization

COPY --from=build /build/initializePlayers ./initializePlayers

CMD [ "/initialization/initializePlayers" ]