FROM golang:1.20

WORKDIR /go/src/gotcp
COPY . .
RUN go mod download
EXPOSE 0-65535
RUN apt-get update
RUN apt-get install -y vim zsh tmux net-tools openssh-server
RUN service ssh start
RUN chsh -s /bin/zsh
RUN sh -c "$(curl -fsSL https://raw.githubusercontent.com/robbyrussell/oh-my-zsh/master/tools/install.sh)"
# RUN go build -o bin/server main/main.go
RUN go build -o bin/server main/main.go
WORKDIR /go/src/gotcp