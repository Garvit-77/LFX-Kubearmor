![image](https://github.com/user-attachments/assets/9e010a50-6416-41fc-a9aa-f6d65c0b9af3)

# KubeArmor Prerequisite Tasks
## Non K8s KubeArmor Enhancements

### 1. Setup KubeArmor in Unorchestrated mode on a BPF LSM node - (https://docs.kubearmor.io/kubearmor/quick-links/kubearmor_vm)
  Challenges : As Kubearmor uses eBPF for monitoring calls and enforcing policies, so there's a need to install BCC (BPF Compiler Collection) tools on your system. </br>
  >Below command would **install the BCC(BPF Complier Collection)**</br>
        `sudo apt-get install bpfcc-tools linux-headers-$(uname -r)`

