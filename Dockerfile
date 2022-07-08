FROM alpine

COPY . /app/

WORKDIR /app

ENTRYPOINT ["/app/main"]


