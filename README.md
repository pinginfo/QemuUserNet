# QemuUserNet

QemuUserNet, crafted in Go, stands as a lightweight solution tailored for QEMU environments. Within its single binary reside both daemon and client functionalities, offering streamlined networking capabilities akin to 'docker network,' but fine-tuned specifically for QEMU. This application operates within user space, utilizing datagram sockets for efficient network management. Its core objective is enabling seamless communication solely between VMs within a virtual network, with no provision for internet connectivity.

## Dependencies

Before you begin, make sure you have the following dependencies installed:

- **Go**: You can download and install it from [the official Go website](https://golang.org/doc/install).

Go will automatically download and install dependencies when you call `go mod download`.

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

## Documentation

To generate documentation for this project, you can use `godoc`. Follow these steps:

1. Make sure you have Go installed on your system. If not, you can download and install it from [the official Go website](https://golang.org/doc/install).
2. Navigate to QemuUserNet directory.
3. Run the following command to generate documentation:
```bash
   godoc -http :6060
```
4. Open your web browser and navigate to `http://localhost:6060/pkg/QemuUserNet`.
    You should see the documentation for your project displayed in your web browser.

5. For more advanced usage of godoc, you can refer to [the official documentation](https://pkg.go.dev/golang.org/x/tools/cmd/godoc).
