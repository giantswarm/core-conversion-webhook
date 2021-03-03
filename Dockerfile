FROM alpine:3.13.2
WORKDIR /app
COPY core-conversion-webhook /app
CMD ["/app/core-conversion-webhook"]
