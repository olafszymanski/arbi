FROM golang:1.18-alpine AS build

RUN mkdir /src

WORKDIR /src

COPY . .

RUN go mod download

RUN go build -o arbi



FROM alpine:latest

RUN mkdir config

COPY --from=build /src/arbi .

COPY --from=build /src/config/config.yml ./config

COPY --from=build /src/credentials.json .

CMD ./arbi