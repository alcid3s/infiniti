# syntax=docker/dockerfile:1

#Constant arguments for the build.
ARG GO_VERSION=1.22.2
ARG PORT=9000

FROM --platform=$BUILDPLATFORM golang:${GO_VERSION} AS final

ENV MAINPATH=/src

WORKDIR ${MAINPATH}

COPY go.mod go.sum ./
RUN go mod download -x

COPY . .

WORKDIR /src/cmd/main

RUN CGO_ENABLED=0 GOOS=linux go build -o /src/bin/infiniti

EXPOSE ${PORT}

CMD ["/src/bin/infiniti"]