# QemuUserNet

QemuUserNet, crafted in Go, stands as a lightweight solution tailored for QEMU environments. Within its single binary reside both daemon and client functionalities, offering streamlined networking capabilities akin to 'docker network,' but fine-tuned specifically for QEMU. This application operates within user space, utilizing datagram sockets for efficient network management. Its core objective is enabling seamless communication solely between VMs within a virtual network, with no provision for internet connectivity.

## Usage

```
Usage: ./QemuUserNet <subcommand> [options]

Subcommands:
  daemon        Start the daemon
  create        Create a network
  connect       Connect a vm to a network
  disconnect    Disconnect a vm to a network
  inspect       Display detailed information on one or more networks
  ls            List networks
  prune         Remove all unused networks
  rm            Remove one or more networks

Options:
  -h string
        Set hostname (default "0.0.0.0")
  -p int
        Set port (default 9000)
```
