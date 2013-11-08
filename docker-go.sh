# See https://github.com/marksteve/docker-go
sudo docker run -i -t -v $GOPATH:/go -w /go/src/github.com/marksteve/bingo go $@
