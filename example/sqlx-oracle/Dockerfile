FROM oraclelinux:8-slim

RUN microdnf install -y wget unzip libaio golang git && \
    microdnf clean all

WORKDIR /tmp
RUN wget https://download.oracle.com/otn_software/linux/instantclient/instantclient-basic-linuxx64.zip && \
    unzip instantclient-basic-linuxx64.zip && \
    mkdir -p /opt && \
    mv instantclient_* /opt/instantclient && \
    rm -f instantclient-basic-linuxx64.zip

WORKDIR /app

COPY . .

ENV LD_LIBRARY_PATH=/opt/instantclient
RUN CGO_ENABLED=1 go build -o app main.go

CMD ["./app"]