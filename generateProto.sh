#!/bin/zsh

## old style to generate grpc code
##protoc greet/greetpb/greet.proto --go_out=plugins=grpc:.

# to generate grpc code
protoc greet/greetpb/greet.proto --go-grpc_out=.
protoc blog/blogpb/blog.proto --go-grpc_out=.

## to generate proto messages
protoc greet/greetpb/greet.proto --go_out=.
protoc blog/blogpb/blog.proto --go_out=.

