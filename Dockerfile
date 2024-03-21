# # Start from a Golang image
# FROM golang:latest

# # Set the Current Working Directory inside the container
# WORKDIR /app

# # Copy go mod and sum files
# COPY go.mod go.sum ./

# # Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
# RUN go mod download

# # Copy the source code from the current directory to the Working Directory inside the container
# COPY . .

# # Build the Go app
# RUN go build -o main .

# # Expose port 8080 to the outside world
# EXPOSE 4000

# # Command to run the executable
# CMD ["./main"]


FROM golang:1.22.1-alpine
ENV GO111MODULE=on

RUN mkdir /app
ADD . /app
WORKDIR /app
RUN apk add git

# Download necessary Go modules
COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

# COPY *.go ./

RUN go build -o /main .
EXPOSE 4000
CMD ["/main"]

