FROM alpine:3.7
ARG BINARY
RUN apk --no-cache add ca-certificates tzdata
RUN cp /usr/share/zoneinfo/Europe/Stockholm /etc/localtime
RUN echo "Europe/Stockholm" >  /etc/timezone
RUN mkdir /app
COPY ${BINARY} /app/pairing-manager
WORKDIR /app
ENTRYPOINT ["./pairing-manager"]