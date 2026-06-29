# Stage 1: Base image for downloading dependencies
FROM golang:alpine AS base

# Set the working directory inside the container
WORKDIR /app

# 设置 Go 模块代理为腾讯云镜像
ENV GOPROXY=https://mirrors.tencent.com/go/,direct

# Copy all file
COPY . .

# Download dependencies
RUN go mod download

# Stage 2: Build stage for each module
FROM base AS builder

ARG MODULE
RUN if [ -z "${MODULE}" ]; then echo "Error: MODULE argument is required." && exit 1; fi

# Build the module
RUN cd ${MODULE} && go build -o ${MODULE}

# Final stage
FROM golang:alpine

WORKDIR /app

ARG MODULE
RUN if [ -z "${MODULE}" ]; then echo "Error: MODULE argument is required." && exit 1; fi
ENV MODULE=${MODULE}

# Copy only the binary and trpc_go.yaml from the builder stage
COPY --from=builder /app/${MODULE}/${MODULE} .
COPY --from=builder /app/${MODULE}/trpc_go.yaml .

# Command to run the application
CMD sh -c "./${MODULE}"