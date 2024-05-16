FROM golang:1.22

WORKDIR /app

COPY . .

ENV CGO_ENABLED=0
RUN go build -o /bin/dnscan ./cmd

FROM scratch

COPY --from=0 /bin/dnscan /bin/dnscan

EXPOSE 9990

ENTRYPOINT ["dnscan"]
