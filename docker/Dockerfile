## WORK IN PROGRESS
## TODO: Please complete the Dockerfile file
## Please refer to the Dockerfile.all file for the final image

# -------------------
# Stage 1: Build Go binaries
# -------------------
FROM golang:1.24 AS builder

WORKDIR /app
COPY . .

RUN cd opengin/core-api && go mod download
RUN cd opengin/core-api && go build ./...
RUN cd opengin/core-api && go build -o core-service cmd/server/service.go cmd/server/utils.go

RUN mkdir -p /app/testbin
RUN cd opengin/core-api/cmd/server && go test -c -o /app/testbin/core-test .
RUN cd opengin/core-api/db/repository/mongo && go test -c -o /app/testbin/mongo-test .
RUN cd opengin/core-api/db/repository/neo4j && go test -c -o /app/testbin/neo4j-test .

# -------------------
# Stage 2: Final Image
# -------------------
FROM ubuntu:22.04

# Install system packages
RUN apt-get update && apt-get install -y \
    curl gnupg lsb-release wget net-tools nano \
    apt-transport-https software-properties-common unzip \
    openjdk-17-jdk openjdk-17-jre \
    && rm -rf /var/lib/apt/lists/*

# Install MongoDB 6.0
RUN curl -fsSL https://pgp.mongodb.com/server-6.0.asc | gpg --dearmor -o /usr/share/keyrings/mongodb-server-6.0.gpg \
    && echo "deb [ signed-by=/usr/share/keyrings/mongodb-server-6.0.gpg ] https://repo.mongodb.org/apt/ubuntu jammy/mongodb-org/6.0 multiverse" | tee /etc/apt/sources.list.d/mongodb-org-6.0.list \
    && apt-get update \
    && apt-get install -y mongodb-org \
    && rm -rf /var/lib/apt/lists/*

# Install Eclipse Temurin JDK 21
RUN wget -O - https://packages.adoptium.net/artifactory/api/gpg/key/public | apt-key add - \
    && echo "deb https://packages.adoptium.net/artifactory/deb $(awk -F= '/^VERSION_CODENAME/{print$2}' /etc/os-release) main" | tee /etc/apt/sources.list.d/adoptium.list \
    && apt-get update \
    && apt-get install -y temurin-21-jdk \
    && rm -rf /var/lib/apt/lists/*

# Set Java environment
ENV JAVA_HOME=/usr/lib/jvm/java-17-openjdk-amd64
ENV PATH=$JAVA_HOME/bin:$PATH

# Verify Java installation
RUN java -version \
    && javac -version \
    && echo "JAVA_HOME: $JAVA_HOME" \
    && ls -la /usr/lib/jvm/ \
    && test -f $JAVA_HOME/bin/java \
    && test -f $JAVA_HOME/bin/javac

# Install Ballerina 2201.8.0 (compatible with Java 17)
RUN wget https://dist.ballerina.io/downloads/2201.8.0/ballerina-2201.8.0-swan-lake.zip \
    && unzip ballerina-2201.8.0-swan-lake.zip \
    && mv ballerina-2201.8.0-swan-lake /usr/lib/ballerina \
    && ln -s /usr/lib/ballerina/bin/bal /usr/bin/bal \
    && rm ballerina-2201.8.0-swan-lake.zip

# Install Neo4j 5.13
RUN wget -O - https://debian.neo4j.com/neotechnology.gpg.key | gpg --dearmor -o /usr/share/keyrings/neo4j.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/neo4j.gpg] https://debian.neo4j.com stable 5" | tee /etc/apt/sources.list.d/neo4j.list \
    && apt-get update \
    && apt-get install -y neo4j=1:5.13.0 cypher-shell \
    && mkdir -p /var/lib/neo4j/data /var/log/neo4j

# Neo4j configuration
RUN sed -i 's/#server.default_listen_address=0.0.0.0/server.default_listen_address=0.0.0.0/' /etc/neo4j/neo4j.conf \
    && sed -i 's/#server.bolt.enabled=true/server.bolt.enabled=true/' /etc/neo4j/neo4j.conf \
    && sed -i 's/#server.bolt.address=0.0.0.0:7687/server.bolt.address=0.0.0.0:7687/' /etc/neo4j/neo4j.conf \
    && sed -i 's/#server.http.enabled=true/server.http.enabled=true/' /etc/neo4j/neo4j.conf \
    && sed -i 's/#server.http.address=0.0.0.0:7474/server.http.address=0.0.0.0:7474/' /etc/neo4j/neo4j.conf \
    && sed -i 's/#dbms.security.auth_enabled=true/dbms.security.auth_enabled=true/' /etc/neo4j/neo4j.conf \
    && echo "dbms.security.procedures.unrestricted=apoc.*" >> /etc/neo4j/neo4j.conf

# Copy compiled binaries and source code
COPY --from=builder /app/opengin/core-api/core-service /usr/local/bin/
COPY --from=builder /app/testbin/* /usr/local/bin/
COPY --from=builder /app/opengin/core-api /app/opengin/core-api
COPY --from=builder /app/opengin/ingestion-api /app/opengin/ingestion-api

WORKDIR /app

# Environment variables
ENV NEO4J_URI=bolt://localhost:7687
ENV NEO4J_USER=neo4j
ENV NEO4J_PASSWORD=neo4j123
ENV MONGO_URI=mongodb://localhost:27017
ENV MONGO_DB_NAME=testdb
ENV MONGO_COLLECTION=metadata

# Expose ports
EXPOSE 7474 7687 27017

# Add entrypoint script
RUN echo '#!/bin/bash\n\
set -e\n\
\n\
NEO4J_PASSWORD=neo4j123\n\
\n\
echo "Starting MongoDB..."\n\
mongod --fork --logpath /var/log/mongodb/mongod.log\n\
\n\
echo "Starting Neo4j..."\n\
neo4j start\n\
\n\
until curl -s http://localhost:7474 > /dev/null; do\n\
    echo "Waiting for Neo4j..."\n\
    sleep 2\n\
done\n\
\n\
echo "Setting Neo4j password..."\n\
echo "ALTER CURRENT USER SET PASSWORD FROM '\''neo4j'\'' TO '\''$NEO4J_PASSWORD'\'';" | cypher-shell -u neo4j -p '\''neo4j'\'' -d system\n\
\n\
until mongosh --eval "db.version()" > /dev/null 2>&1; do\n\
    echo "Waiting for MongoDB..."\n\
    sleep 2\n\
done\n\
\n\
echo "Running CORE service tests..."\n\
cd /app/opengin/core-api\n\
core-test -test.v && mongo-test -test.v && neo4j-test -test.v\n\
\n\
echo "Starting CORE server..."\n\
./core-service &\n\
CORE_PID=$!\n\
sleep 5\n\
\n\
echo "Running ingestion-api tests..."\n\
cd /app/opengin/ingestion-api\n\
bal test\n\
\n\
echo "Stopping CORE server..."\n\
kill $CORE_PID\n\
\n\
tail -f /dev/null' > /start.sh && chmod +x /start.sh

CMD ["/start.sh"]
    