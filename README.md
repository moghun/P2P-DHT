# DHT-14: VoidPhone DHT Module

DHT-14 is a peer-to-peer (P2P) distributed hash table (DHT) implementation based on the Kademlia algorithm. The project is essentially a KVs designed to allow multiple peers to store and retrieve values with 256-bit keys.

## Project Structure

The project consists of several modules, each of which serves a specific role in implementing the P2P DHT system:

### Main Modules:

- **cmd/**: Contains the entry point for both the DHT node and the bootstrap node.
  - **node/**: This folder contains the entry point for running the DHT node (`main.go`).
  - **bootstrap_node/**: Contains the entry point for running the bootstrap node (`main.go`).
- **pkg/**: Contains the core logic of the project, split into various components.

  - **node/**: Implements the logic for node management, including bootstrap and liveliness checks.
  - **dht/**: Contains the Kademlia-based Distributed Hash Table implementation, including key handling, replication, routing table, and success message responses.
  - **api/**: Implements the API server, handlers, middleware
  - **message/**: Defines the various message types used for communication between nodes, such as PUT, GET, and PING messages.
  - **security/**: Implements various security functionalities for P2P application namely encryption module for the secure local data storage, rate limiter, Proof-of-Work mechanism and self-signed TLS implementation.
  - **storage/**: Local encrypted KV storage implementation.
  - **util/**: Provides utility functions like configuration parsing, logging and encryption.

- **tests/**: Contains the unit and integration tests.
  - **integration_tests/**: End-to-end tests for verifying the integration of various components.
  - **unit_tests/**: Contains unit tests for individual components of the project.

## Technologies Used

This project is built using the following technologies:

- **Go**: The main language used for building the DHT.
- **gotest**: The testing framework used for writing and executing unit and integration tests.
- **Docker**: Containerization of the DHT nodes and bootstrap nodes for easier deployment.
- **Make**: The build system used to automate various tasks like building, testing, and running the project.
- **Logrus**: A structured logger used for logging within the application.

## Instructions

### Configuration

The application is configured using a `config.ini` file. Below is an example of a configuration file:

```ini
[node]
p2p_address = 0.0.0.0:8000
api_address = 0.0.0.0:7400
cleanup_interval = 86400
bootstrap_retry_interval = 5
max_bootstrap_retries = 2

[bootstrap]
bootstrap_node_1 = bootstrap_node_1:8000
bootstrap_node_2 = bootstrap_node_2:8000

[security]
encryption_key = 12345678901234567890123456789012
difficulty = 4

[rate_limiter]
requests_per_second = 100
burst_size = 200
```

## Usage

## Auto Local Deployment:

To deploy a local system with multiple nodes, use the `./deploy.sh` script. The script handles building Docker containers, creating a Docker network, and starting both bootstrap and DHT nodes.

1. **Make the script executable**:
   ```bash
   chmod +x deploy.sh
   ```
2. **Deploy nodes**:
   ```bash
   ./deploy.sh <number_of_nodes>

   ```
   This builds the Docker image, creates the dht-14_network, starts 2 bootstrap nodes, and deploys the specified number of DHT nodes.

   ```

3. **Stop and remove running containers:**

   ```bash
   ./deploy.sh stop

   ```

## Example Run Using Client:

Once the nodes are running, you can use the client_dht.py script to interact with the DHT. Below is a step-by-step guide:

1. Deploy the nodes using ./deploy.sh or manually start one bootstrap node and at least one DHT node.

2. Navigate to the dummy_client directory:

```bash
cd dummy_client

```

3. Set up a Python virtual environment:

```bash
python3 -m venv {myenvname}
source {myenvname}/bin/activate
```

4. Install the required Python packages:

```bash
pip install -r requirements.txt
```

**If using the ./deploy.sh script, the API ports for the nodes will start at 7401 and increase sequentially**

5. To store (PUT) a value in the DHT:

```bash
python3 dht_client.py -a 127.0.0.1 -p {node_api_port} -s
```

6. To retrieve (GET) the value from the DHT:

```bash
python3 dht_client.py -a 127.0.0.1 -p {node_api_port} -g
```

## Running the Application Manually

#### Run a DHT Node

You can manually run a DHT node using the following Makefile target:

```bash
make run_node
```

#### Run a Bootstrap Node

Similarly, to run a bootstrap node, use:

```bash
make run_bootstrap
```

#### Running Tests

You can run unit and integration tests with the following command:

```bash
make test
```

#### Generate HTML Test Report

To generate a test report in HTML format, use:

```bash
make test_and_report
```

## Docker Usage:

#### Building the Docker Image

To build the Docker image:

```bash
docker build -t {arbitrary_image_name}:latest .
```

#### Creating a Docker Network

Create a Docker network for the nodes to communicate:

```bash
docker network create {arbitrary_p2p_network_name}
```

#### Running a DHT Node

To run a DHT node in a Docker container:

```bash
docker run -d \
  --name node_{node_name} \
  --network {arbitrary_p2p_network_name} \
  -p {node_p2p_port}:8000 -p {node_api_port}:7400 \
  -v $(pwd)/config.ini:/app/config.ini \
  -e NODE_TYPE=node \
  {arbitrary_image_name}:latest
```

#### Running a Bootstrap Node

To run a bootstrap node in a Docker container:

```bash
docker run -d \
  --name bootstrap_node_{node_name} \
  --network {arbitrary_p2p_network_name} \
  -p {bootstrap_node_p2p_port}:8000 \
  -v $(pwd)/config.ini:/app/config.ini \
  -e NODE_TYPE=bootstrap_node \
  {arbitrary_image_name}:latest
```

## Makefile Commands

- `build`: Build both the DHT node and the bootstrap node.
- `run_node`: Run the DHT node.
- `run_bootstrap`: Run the bootstrap node.
- `test`: Run unit and integration tests.
- `test_and_report`: Run tests and generate an HTML report.
- `clean`: Clean up build artifacts and test results.

## Dockerfile Details

The Dockerfile is a two-stage build that first compiles the Go binaries and then creates a minimal Alpine Linux container for running the DHT node or bootstrap node. The correct binary is chosen at runtime based on the `NODE_TYPE` environment variabled.

- **Stage 1**: The application is built using Go.
- **Stage 2**: A minimal container with the statically linked binaries is created.

## Deploy Script

The `deploy.sh` script handles the deployment of both the bootstrap nodes and DHT nodes. It supports stopping and cleaning up existing containers as well.

- **Inputs**: Number of DHT nodes to deploy.
- **Outputs**: Running Docker containers for the bootstrap and DHT nodes.

### Usage:

```bash
./deploy.sh <number_of_nodes>
./deploy.sh stop
```
