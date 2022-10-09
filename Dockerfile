FROM zenika/alpine-chrome:latest

COPY --from=golang:1.19-alpine /usr/local/go/ /usr/local/go/
ENV PATH="/usr/local/go/bin:${PATH}"

COPY . /server
WORKDIR /server
USER root
RUN chmod 777 -R /tmp
RUN go build
ENTRYPOINT [ "/server/hangang-view-server" ]