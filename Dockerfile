FROM golang:alpine as build-env

ENV GO111MODULE=on
RUN apk --no-cache add git

COPY . /app
WORKDIR /app

RUN mkdir /recordings

RUN ls -lahR && go mod download && go build -o /vnc-recorder-server

FROM jrottenberg/ffmpeg:4.1-alpine
COPY --from=build-env /vnc-recorder-server /
ENTRYPOINT ["/vnc-recorder-server"]
CMD [""]
