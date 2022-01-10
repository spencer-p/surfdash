FROM golang:1.17 as builder

WORKDIR /go/src/github.com/spencer-p/surfdash
COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY . ./
RUN CGO_ENABLED=0 go install .

FROM gcr.io/distroless/static-debian11
COPY --from=builder /go/bin/surfdash /app

ENV TZ=America/Los_Angeles
EXPOSE 8080
CMD ["/app"]
