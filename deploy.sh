#!/bin/bash

# Function to remove exited containers
cleanup_exited_containers() {
  echo "Cleaning up exited containers..."
  exited_containers=$(docker ps -a -q --filter "status=exited")

  if [ -n "$exited_containers" ]; then
    docker rm $exited_containers
    echo "Exited containers removed."
  else
    echo "No exited containers to remove."
  fi
}

# Function to stop and remove all DHT containers
stop_and_remove_containers() {
  echo "Stopping and removing all DHT containers..."
  
  # List of DHT containers to stop and remove (based on names)
  container_names=$(docker ps -a --filter "name=bootstrap_node_1" --filter "name=bootstrap_node_2" --filter "name=node_" --format "{{.Names}}")

  if [ -z "$container_names" ]; then
    echo "No DHT containers found."
  else
    # Stop and remove the containers
    docker stop $container_names
    docker rm $container_names
    echo "All DHT containers have been stopped and removed."
  fi
}

# Function to generate a unique build name
generate_unique_name() {
  local unique_id=$(date +"%Y%m%d%H%M%S")
  echo "dht-14_${unique_id}_project"
}

# Function to deploy bootstrap nodes
deploy_bootstrap_nodes() {
  local build_name=$1

  echo "Deploying bootstrap nodes with image name: $build_name..."
  
  # Create the Docker network if it doesn't exist
  if [ -z $(docker network ls --filter name=dht-14_network --format "{{.Name}}") ]; then
    docker network create dht-14_network
    echo "Docker network dht-14_network created."
  else
    echo "Docker network dht-14_network already exists."
  fi

  # Deploy the bootstrap nodes
  docker run -d \
    --name bootstrap_node_1 \
    --network dht-14_network \
    -p 8081:8000 \
    -v $(pwd)/config.ini:/app/config.ini \
    -e NODE_TYPE=bootstrap_node \
    "$build_name"

  docker run -d \
    --name bootstrap_node_2 \
    --network dht-14_network \
    -p 8082:8000 \
    -v $(pwd)/config.ini:/app/config.ini \
    -e NODE_TYPE=bootstrap_node \
    "$build_name"

  # Wait for 2 seconds to allow the bootstrap nodes to start
  echo "Waiting for bootstrap nodes to start..."
  sleep 2
}

# Function to deploy DHT nodes
deploy_dht_nodes() {
  local node_count=$1
  local build_name=$2

  echo "Deploying $node_count DHT nodes with image name: $build_name..."

  for ((i=1; i<=node_count; i++))
  do
    node_port=$((8000 + i))
    api_port=$((7400 + i))
    node_name="node_$i"

    echo "Deploying $node_name with P2P port $node_port and API port $api_port"

    docker run -d \
      --name "$node_name" \
      --network dht-14_network \
      -p "$node_port":8000 \
      -p "$api_port":7400 \
      -v $(pwd)/config.ini:/app/config.ini \
      -e NODE_TYPE=node \
      "$build_name"
  done
}

# Main execution
if [ "$#" -lt 1 ]; then
  echo "Usage: $0 <number_of_nodes> OR $0 stop"
  exit 1
fi

if [ "$1" == "stop" ]; then
  stop_and_remove_containers
  exit 0
fi

node_count=$1
build_name=$(generate_unique_name)

# Clean up any exited containers
cleanup_exited_containers

# Build the Docker image with a unique name
docker build -t "$build_name" .

# Deploy bootstrap nodes
deploy_bootstrap_nodes "$build_name"

# Deploy DHT nodes
deploy_dht_nodes "$node_count" "$build_name"