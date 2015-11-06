FROM golang:1.5.1
COPY . /usr/go/src/arrows.io/cap
WORKDIR /usr/go/src/arrows.io/cap
ENV GOPATH /usr/go/
RUN go get -d -v
RUN go install -v
EXPOSE 8080
ENTRYPOINT ["/usr/go/bin/cap"]