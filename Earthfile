VERSION 0.7
FROM golang:1.21.5-alpine3.19
WORKDIR /exo

test:
	COPY go.mod ./
	COPY exo.go ./
	COPY --dir changeset ./
	RUN go test
	RUN go test ./changeset

build:
	COPY go.mod ./
	COPY exo.go ./
	COPY --dir changeset ./
	RUN go build
	RUN go build ./changeset
