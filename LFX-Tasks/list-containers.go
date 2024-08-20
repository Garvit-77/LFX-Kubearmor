package main

import (
    "context"
    "fmt"
    "log"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/client"
)

func main() {
    // Create a Docker client
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        log.Fatalf("Error creating Docker client: %v", err)
    }

    // List containers
    containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
    if err != nil {
        log.Fatalf("Error listing containers: %v", err)
    }

    // Print container details
    for _, container := range containers {
        fmt.Printf("Container ID: %s\n", container.ID)
        fmt.Printf("Image: %s\n", container.Image)
        fmt.Printf("Status: %s\n", container.Status)
        fmt.Printf("Names: %v\n", container.Names)
        fmt.Println("-----------")
    }
}

