FROM alpine:3.7
ARG BINARY
RUN apk --no-cache add ca-certificates
RUN mkdir /app
COPY ${BINARY} /app/pairing-manager
WORKDIR /app
ENTRYPOINT ["./pairing-manager"]