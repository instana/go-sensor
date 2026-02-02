FROM oraclelinux:8-slim

RUN microdnf install -y wget unzip libaio golang git && \
    microdnf clean all

WORKDIR /tmp
RUN wget --max-redirect=0 https://download.oracle.com/otn_software/linux/instantclient/instantclient-basic-linuxx64.zip && \
    unzip instantclient-basic-linuxx64.zip && \
    mkdir -p /opt && \
    mv instantclient_* /opt/instantclient && \
    rm -f instantclient-basic-linuxx64.zip

WORKDIR /app

# Copy only necessary files for the build
COPY go.mod go.sum ./
RUN go mod download

COPY main.go ./

ENV LD_LIBRARY_PATH=/opt/instantclient
RUN CGO_ENABLED=1 go build -o app main.go

# Create a non-root user to run the application
RUN groupadd -r appuser && useradd -r -g appuser appuser && \
    chown -R appuser:appuser /app && \
    chmod -R 755 /app

USER appuser

CMD ["./app"]
