FROM golang:1.18-alpine AS build

RUN mkdir /src

WORKDIR /src

COPY . .

RUN go mod download

RUN go build -o arbi



FROM alpine:latest

COPY --from=build /src/arbi .

CMD ./arbi

