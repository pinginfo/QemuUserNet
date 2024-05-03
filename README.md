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

## Subcommand

### Start the daemon
```
Usage: ./QemuUserNet daemon [options]

Options:
  -h string
        Set hostname (default "0.0.0.0")
  -p int
        Set port (default 9000)
```

### Create a network
```
Usage: ./QemuUserNet create [options] NETWORK

Options:
  -h string
        Set hostname (default "0.0.0.0")
  -p int
        Set port (default 9000)
  -subnet string
        Subnet in CIDR format that represents a network segment (default "10.10.10.0/24")
```

### Connect to a network
```
Usage: ./QemuUserNet connect [options] NETWORK MAC

Options:
  -h string
        Set hostname (default "0.0.0.0")
  -p int
        Set port (default 9000)
```

### Disconnect to a network
```
Usage: ./QemuUserNet disconnect [options] NETWORK MAC

Options:
  -h string
        Set hostname (default "0.0.0.0")
  -p int
        Set port (default 9000)
```

### Inspect a network
```
Usage: ./QemuUserNet inspect [options] NETWORK [NETWORK...]

Options:
  -h string
        Set hostname (default "0.0.0.0")
  -p int
        Set port (default 9000)
```

### List networks
```
Usage: ./QemuUserNet ls [options]

Options:
  -h string
        Set hostname (default "0.0.0.0")
  -p int
        Set port (default 9000)
```

### Remove all unused networks
```
Usage: ./QemuUserNet prune [options]

Options:
  -h string
        Set hostname (default "0.0.0.0")
  -p int
        Set port (default 9000)
```

### Remove networks
```
Usage: ./QemuUserNet rm [options] NETWORK [NETWORK...]

Options:
  -h string
        Set hostname (default "0.0.0.0")
  -p int
        Set port (default 9000)
```
