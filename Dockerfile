FROM golang:1.10 as builder

RUN go get -u github.com/golang/dep/cmd/dep

WORKDIR $GOPATH/src/github.com/fnproject/hotwrap

COPY Gopkg.toml Gopkg.lock ./

RUN dep ensure --vendor-only

COPY . ./
 
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="-w -s" -o  /hotwrap  $GOPATH/src/github.com/fnproject/hotwrap/hotwrap.go

FROM scratch

COPY --from=builder /hotwrap /hotwrap