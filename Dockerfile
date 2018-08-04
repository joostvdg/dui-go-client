FROM golang:latest AS build
WORKDIR /src
ENV LAST_UPDATE=20180209
RUN go get -v github.com/gorilla/mux/...
#RUN go get -v github.com/stretchr/testify/...
ADD . /src
RUN go test --cover ./...
RUN go build -v -tags netgo -o dui-go

FROM alpine:3.8
ENV LAST_UPDATE=20180506
LABEL authors="Joost van der Griendt <joostvdg@gmail.com>"
LABEL version="0.1.0"
LABEL description="Docker image for participating in DUI"
EXPOSE 7777
CMD ["dui-go"]
# HEALTHCHECK --interval=5s --start-period=3s --timeout=5s CMD wget -qO- "http://localhost:7777/stacks"
COPY --from=build /src/dui-go /usr/local/bin/dui-go
RUN chmod +x /usr/local/bin/dui-go