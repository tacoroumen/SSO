# Use the official Go image as the base image
FROM golang:1.18

# Set the working directory inside the container
WORKDIR /go/src/app

# Copy the Go source code and necessary files into the container
COPY . .

# Build the Go application
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the Go application
CMD ["./main"]
