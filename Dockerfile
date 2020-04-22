FROM golang:alpine AS builder

RUN apk update && apk add --no-cache git

WORKDIR $GOPATH/src/browserstack-prometheus-exporter/
COPY plan.go .

RUN go get -d -v
RUN GOOS=linux CGO_ENABLED=0 go build -a -o /go/bin/plan plan.go


FROM scratch

COPY --from=builder /go/bin/plan /go/bin/plan

EXPOSE 5123

ENTRYPOINT ["/go/bin/plan"]
