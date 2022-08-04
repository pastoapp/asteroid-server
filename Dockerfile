FROM golang:1.18-alpine

ARG IPFS_URL="http://ipfs:5001"
ARG ORBIT_DB_DIR="/data/orbitdb"

# OrbitDB Location
ENV ORBIT_DB_LOCATION $ORBIT_DB_DIR

# IPFS HTTP API
ENV IPFS_API_URL $IPFS_URL

# Set the Current Working Directory inside the container
WORKDIR $GOPATH/src/

# Copy everything from the current directory to the PWD (Present Working Directory) inside the container
COPY . .

# Download all the dependencies
RUN go get -d -v ./...

# Install the package
RUN go install -v ./...

# Create OrbitDB Directory
RUN mkdir -p $ORBIT_DB_LOCATION

# This container exposes port 8080 to the outside world
EXPOSE 3000

# Run the executable
CMD ["asteroid-api", "-orbitdb-dir", "$ORBIT_DB_LOCATION", "-ipfs-url", "$IPFS_API_URL"]
