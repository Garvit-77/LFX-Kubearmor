package main

import (
    "context"
    "fmt"
    "log"
    "os"
    "os/exec"
    "path/filepath"
    "text/template"

    "github.com/docker/docker/api/types"
    "github.com/docker/docker/client"
)

// KubeArmorPolicy defines the structure for the YAML file
type KubeArmorPolicy struct {
    ContainerName string
}

// Template for the KubeArmorPolicy YAML
const kubeArmorPolicyTemplate = `apiVersion: security.kubearmor.com/v1
kind: KubeArmorPolicy
metadata:
  name: process-block-{{ .ContainerName }}
spec:
  severity: 5
  message: "a critical file was accessed"
  tags:
  - WARNING
  selector:
    matchLabels:
      kubearmor.io/container.name: {{ .ContainerName }}
  process:
    matchPaths:
      - path: /usr/bin/ls
      - path: /usr/bin/sleep
  action:
    Block
`

func main() {
    cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
    if err != nil {
        log.Fatalf("Error creating Docker client: %v", err)
    }

    containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
    if err != nil {
        log.Fatalf("Error listing containers: %v", err)
    }

    for _, container := range containers {
        // Use the first name of the container as the container name
        containerName := container.Names[0][1:] // remove the leading '/'
        policy := KubeArmorPolicy{ContainerName: containerName}

        // Generate YAML file
        yamlFileName := fmt.Sprintf("kubearmor_policy_%s.yaml", containerName)
        err := generateYAMLFile(yamlFileName, policy)
        if err != nil {
            log.Fatalf("Error generating YAML file: %v", err)
        }

        // Apply the KubeArmorPolicy using the karmor command
        err = applyKubeArmorPolicy(yamlFileName)
        if err != nil {
            log.Fatalf("Error applying KubeArmor policy: %v", err)
        }

        fmt.Printf("Applied KubeArmorPolicy for container: %s\n", containerName)
    }
}

func generateYAMLFile(fileName string, policy KubeArmorPolicy) error {
    tmpl, err := template.New("policy").Parse(kubeArmorPolicyTemplate)
    if err != nil {
        return err
    }

    file, err := os.Create(fileName)
    if err != nil {
        return err
    }
    defer file.Close()

    return tmpl.Execute(file, policy)
}

func applyKubeArmorPolicy(yamlFileName string) error {
    absPath, err := filepath.Abs(yamlFileName)
    if err != nil {
        return err
    }

    cmd := exec.Command("karmor", "vm", "policy", "add", absPath)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return fmt.Errorf("error running karmor: %v, output: %s", err, output)
    }

    return nil
}
