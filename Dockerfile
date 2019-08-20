FROM golang:alpine as build

RUN apk --update add git upx

RUN adduser -h /app -D kubeconsole

# Setup work env
RUN mkdir -p /app /build/gocode
ADD . /app/
WORKDIR /app

# Required envs for GO
ENV GOPATH=/build/gocode
ENV GOOS=linux
ENV GOARCH=amd64
ENV GO111MODULE=on

# Disable CGO so we can use a scratch container
ENV CGO_ENABLED=0

RUN cd cmd/kubeconsole && go build -o /app/kubeconsole .
RUN upx /app/kubeconsole


# Use a scratch container so nothing but the app is present
FROM scratch

# Copy the binary and some directories from build
COPY --from=build /app/kubeconsole /app/kubeconsole
COPY --from=build /tmp /tmp
COPY --from=build /etc/passwd /etc/passwd

USER kubeconsole

# Start the app
ENTRYPOINT ["/app/kubeconsole"]
