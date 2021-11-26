FROM golang:1.17 as builder

WORKDIR /go/src/app
COPY ["go.mod", "go.sum", "./"]
RUN go mod download

COPY * ./
RUN CGO_ENABLED=0 go install .

FROM gcr.io/distroless/static-debian11
COPY --from=builder /go/bin/app /app
COPY kodata /kodata

ENV KO_DATA_PATH=/kodata
EXPOSE 8080
CMD ["/app"]
