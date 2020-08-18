## Description
This is Video Bank Service. This service provide basic CRUD function and some others function.
The data is stored to MongoDB. gRPC is used here for the communication.

Stack Tech: Golang, MongoDB, gRPC, docker


## Clone/Download Repo
git clone https://github.com/kenanya/jt-video-bank.git

## How to Start
cd jt-video-bank/cmd/server
go build .
APP_ENV=local ./server
