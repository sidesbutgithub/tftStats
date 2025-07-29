FROM golang:1.24.5-alpine AS build

WORKDIR /build

COPY go.mod go.mod ./

RUN go mod download

COPY . ./

RUN go build -o ./crawler


FROM gcr.io/distroless/base

WORKDIR /crawler

COPY --from=build /build/crawler ./matchCrawler

CMD [ "/crawler/matchCrawler" ]