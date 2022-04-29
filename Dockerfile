FROM public.ecr.aws/docker/library/golang:1.18 as build-env

WORKDIR /go/src/app

COPY go.* ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 go build -tags=containers_image_openpgp -o /go/bin/vendorito ./cmd/vendorito

FROM public.ecr.aws/docker/library/debian:bullseye-slim

RUN apt-get update \
    && apt-get install -y ca-certificates tzdata \
    && rm -r /var/lib/apt/lists/ /var/cache/apt/archives

COPY --from=build-env /go/bin/vendorito /usr/local/bin

CMD ["vendorito"]