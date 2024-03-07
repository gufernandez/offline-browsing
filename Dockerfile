# Use an official Golang runtime as a parent image
 FROM golang:alpine

 # Set the working directory inside the container
 WORKDIR /app

 # Copy the local package files to the container's workspace
 COPY . /app

 # Build the Go application inside the container
 RUN go install github.com/spf13/cobra-cli@latest
 RUN go build -o fetch
