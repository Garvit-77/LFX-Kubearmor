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

3. **Replace Container Names in KubeArmor Policy**:
   - Modify the Go script to replace the container names in the KubeArmor policy template:
     ```go
     policyTemplate := `apiVersion: security.kubearmor.com/v1
     kind: KubeArmorPolicy
     metadata:
       name: example-container-policy
       annotations:
         kubearmor.io/container.name: lb
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

5. **Automatically Apply Policies**:
   - Extend the script to automatically call the function that applies policies:
     ```go
     func applyPolicy(policy string) {
         // Code to apply policy using karmor CLI
     }
     
     func main() {
         // Code to extract containers and generate policies
         applyPolicy(updatedPolicy)
     }
     ```

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


### Future Work

- **Enhanced Automation**: Consider further automating the policy generation and application process.
- **Policy Optimization**: Explore ways to optimize KubeArmor policies for better performance and security.
