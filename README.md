![image](https://github.com/user-attachments/assets/9e010a50-6416-41fc-a9aa-f6d65c0b9af3)

# KubeArmor Prerequisite Tasks
## Non K8s KubeArmor Enhancements

# KubeArmor in Unorchestrated Mode on a BPF LSM Node

This document outlines the process of setting up KubeArmor in Unorchestrated mode on a BPF LSM node, creating Docker containers, writing a Go script to manage KubeArmor policies, and the challenges faced during the task.

## Prerequisites

- A BPF LSM-enabled Linux system.
- Docker installed and running.
- Go programming environment set up.
- Basic knowledge of KubeArmor and Docker.

## Steps

### 1. Setting Up KubeArmor in Unorchestrated Mode

1. **Follow the KubeArmor Setup Guide**:
   - Use the official KubeArmor documentation [here](https://docs.kubearmor.io/kubearmor/quick-links/kubearmor_vm) to set up KubeArmor in Unorchestrated mode.
   - Ensure that the system is running on a BPF LSM node.

2. **Verify the Installation**:
   - Run `sudo kubectl get pods -n kube-system` to check if KubeArmor is running properly.

#### Challenges Faced

- **Issue with BPF LSM Compatibility**:
  - **Problem**: The BPF LSM node setup was complex, and ensuring compatibility with KubeArmor required additional kernel modules and configurations.
  - **Solution**: Verified the kernel version and ensured BPF LSM support by checking `/sys/kernel/security/lsm`.

### 2. Creating Docker Containers

1. **Create Docker Containers**:
   - Run the following command to create a Docker container:
     ```bash
     docker run -d --name test nginx
     ```
   - Repeat the process to create additional containers if needed.

2. **Verify the Containers**:
   - Check the running containers using:
     ```bash
     docker ps
     ```

#### Challenges Faced

- **Port Conflicts**:
  - **Problem**: Running containers on the default port 8080 conflicted with other services.
  - **Solution**: Modified the Docker command to run containers on different ports or stopped conflicting services.

### 3. Writing a Go Script to Manage KubeArmor Policies

1. **Setup Go Environment**:
   - Ensure that Go is installed and properly set up. You can check the installation with:
     ```bash
     go version
     ```

2. **Using the Docker API to Extract Running Containers**:
   - Write a Go script to interact with the Docker API and retrieve the list of running containers. Below is an example code snippet:
     ```go
     package main

     import (
         "context"
         "fmt"
         "github.com/docker/docker/api/types"
         "github.com/docker/docker/client"
     )

     func main() {
         cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
         if err != nil {
             panic(err)
         }

         containers, err := cli.ContainerList(context.Background(), types.ContainerListOptions{})
         if err != nil {
             panic(err)
         }

         for _, container := range containers {
             fmt.Println(container.Names[0])
         }
     }
     ```
     This Script can be found by name list-containers.go in LFX-Tasks Dir

3. **Replace Container Names in KubeArmor Policy**:
   - Modify the Go script to replace the container names in the KubeArmor policy template:
     ```go
     policyTemplate := `apiVersion: security.kubearmor.com/v1
     kind: KubeArmorPolicy
     metadata:
       name: example-container-policy
       annotations:
         kubearmor.io/container.name: busy_babbage
         kubearmor.io/container.name: ishizaka
     spec:
       process:
         matchPaths:
         - path: /bin/cat
     `
     updatedPolicy := strings.Replace(policyTemplate, "lb", "test", -1)
     fmt.Println(updatedPolicy)
     ```

4. **Apply the Policies Using `karmor`**:
   - Use the `karmor` CLI tool to apply the generated policies:
     ```bash
     karmor vm policy add ./generated_policy.yaml
     ```
     - this created kubearmor_containerpolicy.yaml,kubearmor_policy_busy_babbage.yaml which can be seen in LFX_Tasks Directory

5. **Automatically Apply Policies**:
   - Extend the script to automatically call the function that applies policies:
     ```go
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
     ```

     This code can be found in LFX-Tasks dir named by main.go

#### Challenges Faced

- **Docker API Version Issues**:
  - **Problem**: The Docker API version compatibility with the Go client library caused issues when retrieving container lists.
  - **Solution**: Used `client.WithAPIVersionNegotiation()` in the Go client setup to handle version mismatches automatically.

- **String Replacement Errors**:
  - **Problem**: The container name replacement in the policy sometimes failed due to incorrect string formatting.
  - **Solution**: Ensured proper use of the `strings.Replace()` function and checked for extra spaces or newline characters.

- **Policy Application Failures**:
  - **Problem**: Policies sometimes failed to apply due to syntax errors in the generated YAML.
  - **Solution**: Added additional validation and formatting steps before applying the policy.

### 4. Checking for Policy Violations

1. **Monitor Logs for Violations**:
   - Use `karmor` to monitor logs and check for policy violations:
     ```bash
     karmor vm log
     ```
   - Verify that the policies are being enforced correctly by performing actions in the containers that would trigger a violation.

#### Challenges Faced

- **Log Monitoring Performance**:
  - **Problem**: Monitoring logs in real-time caused performance issues on the node.
  - **Solution**: Limited the number of log entries displayed or used filtering options to focus on specific containers or events.

