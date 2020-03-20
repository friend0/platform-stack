FROM golang:1.12 as go_build

ENV GO111MODULE=on

ARG GIT_TOKEN
RUN git config --global url.https://${GIT_TOKEN}@github.com/.insteadOf https://github.com/

WORKDIR /src/backend

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./...

FROM alpine:3.7

# NOTE: Using 3.7 above to supports psql 10.x below
RUN apk update && apk add postgresql

COPY --from=go_build /src/backend/main /opt/backend/

EXPOSE 8000
WORKDIR /opt/backend

CMD [ "./main" ]
