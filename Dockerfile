FROM golang:1.11

LABEL maintainer="Olivier Sallou <olivier.sallou@irisa.fr>"

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/github.com/osallou/goterra-injector

# Copy everything from the current directory to the PWD(Present Working Directory) inside the container
COPY . .
RUN go get -u github.com/golang/dep/cmd/dep
#RUN go get -d -v ./...
RUN cd tools/linter && dep ensure
# Install the package
RUN go build -o goterra-linter tools/linter/linter.go
RUN cd tools/injector && dep ensure
RUN go build -o goterra-injector tools/injector/goterra-injector.go
RUN cp tools/injector/goterra.yml.example goterra.yml

FROM alpine:latest  
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=0 /go/src/github.com/osallou/goterra-injector/goterra-linter .
COPY --from=0 /go/src/github.com/osallou/goterra-injector/goterra-injector .
COPY --from=0 /go/src/github.com/osallou/goterra-injector/goterra.yml .
RUN mkdir /lib64 && ln -s /lib/libc.musl-x86_64.so.1 /lib64/ld-linux-x86-64.so.2
CMD ["./goterra-injector"]
