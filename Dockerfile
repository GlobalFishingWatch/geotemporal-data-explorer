FROM golang:1.18-stretch
ENV DUCKDB_VERSION=0.4.0
RUN apt-get update -y && apt-get install -y ca-certificates unzip && update-ca-certificates 
WORKDIR /go/src/app
COPY . .
CMD ["sleep", "90000"]