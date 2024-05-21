package main

import (
	"QemuUserNet/client"
	"QemuUserNet/daemon"
	"flag"
	"fmt"
	"log"
	"os"
)

func main() {
	var (
		ip         string
		port       int
		subnet     string
		gatewayIP  string
		gatewayMAC string
		rangeIP    string
	)

	daemonCmd := flag.NewFlagSet("daemon", flag.ExitOnError)
	createCmd := flag.NewFlagSet("create", flag.ExitOnError)
	connectCmd := flag.NewFlagSet("connect", flag.ExitOnError)
	disconnectCmd := flag.NewFlagSet("disconnect", flag.ExitOnError)
	inspectCmd := flag.NewFlagSet("inspect", flag.ExitOnError)
	lsCmd := flag.NewFlagSet("ls", flag.ExitOnError)
	pruneCmd := flag.NewFlagSet("prune", flag.ExitOnError)
	rmCmd := flag.NewFlagSet("rm", flag.ExitOnError)

	createCmd.StringVar(&subnet, "subnet", "10.10.10.0/24", "Subnet in CIDR format that represents a network segment")
	createCmd.StringVar(&gatewayIP, "gateway", "10.10.10.1", "The IP address of the gateway for the network segment")
	createCmd.StringVar(&gatewayMAC, "gatewaymac", "52:54:00:12:34:ff", "The MAC (Media Access Control) address of the gateway device")
	createCmd.StringVar(&rangeIP, "rangeip", "10.10.10.100-200", "A range of IP addresses within the subnet that can be assigned to devices. The range is specified with a start and end IP address, indicating the pool of IP addresses available for DHCP assignment")

	flag.StringVar(&ip, "h", "0.0.0.0", "Set hostname")
	flag.IntVar(&port, "p", 9000, "Set port")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s <subcommand> [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nSubcommands:\n")
		fmt.Fprintf(os.Stderr, "  daemon	Start the daemon\n")
		fmt.Fprintf(os.Stderr, "  create	Create a network\n")
		fmt.Fprintf(os.Stderr, "  connect	Connect a vm to a network\n")
		fmt.Fprintf(os.Stderr, "  disconnect	Disconnect a vm to a network\n")
		fmt.Fprintf(os.Stderr, "  inspect	Display detailed information on one or more networks\n")
		fmt.Fprintf(os.Stderr, "  ls		List networks\n")
		fmt.Fprintf(os.Stderr, "  prune		Remove all unsuned networks\n")
		fmt.Fprintf(os.Stderr, "  rm		Remove one or more networks\n")
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
	}

	daemonCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s daemon [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		daemonCmd.PrintDefaults()
	}

	createCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s create [options] NETWORK\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		createCmd.PrintDefaults()
	}

	connectCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s connect [options] NETWORK MAC\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		connectCmd.PrintDefaults()
	}

	disconnectCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s disconnect [options] NETWORK MAC\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		disconnectCmd.PrintDefaults()
	}

	inspectCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s inspect [options] NETWORK [NETWORK...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		inspectCmd.PrintDefaults()
	}

	lsCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s ls [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		lsCmd.PrintDefaults()
	}

	pruneCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s prune [options]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		pruneCmd.PrintDefaults()
	}

	rmCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s rm [options] NETWORK [NETWORK...]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		rmCmd.PrintDefaults()
	}

	if len(os.Args) < 2 {
		flag.Usage()
		os.Exit(0)
	}

	switch os.Args[1] {
	case "daemon":
		daemonCmd.Parse(os.Args[2:])
		daemon.InitDaemon(ip, port)
	case "create":
		createCmd.Parse(os.Args[2:])
		if createCmd.NArg() != 1 {
			createCmd.Usage()
			os.Exit(0)
		}
		err := client.Create(ip, port, createCmd.Arg(0), subnet, gatewayIP, gatewayMAC, rangeIP)
		if err != nil {
			log.Println("error: ", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	case "connect":
		connectCmd.Parse(os.Args[2:])
		if connectCmd.NArg() != 2 {
			connectCmd.Usage()
			os.Exit(0)
		}
		err := client.Connect(ip, port, connectCmd.Arg(0), connectCmd.Arg(1))
		if err != nil {
			log.Println("error: ", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	case "disconnect":
		disconnectCmd.Parse(os.Args[2:])
		if disconnectCmd.NArg() != 2 {
			disconnectCmd.Usage()
			os.Exit(0)
		}
		err := client.Disconnect(ip, port, disconnectCmd.Arg(0), disconnectCmd.Arg(1))
		if err != nil {
			log.Println("error", err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	case "inspect":
		inspectCmd.Parse(os.Args[2:])
		if inspectCmd.NArg() < 1 {
			inspectCmd.Usage()
			os.Exit(0)
		}
		networkNames := make([]string, inspectCmd.NArg())
		for i := 0; i < inspectCmd.NArg(); i++ {
			networkNames[i] = inspectCmd.Arg(i)
		}
		err := client.Inspect(ip, port, networkNames)
		if err != nil {
			log.Println("error", err.Error())
			os.Exit(1)
		}
	case "ls":
		lsCmd.Parse(os.Args[2:])
		if lsCmd.NArg() != 0 {
			lsCmd.Usage()
			os.Exit(0)
		}
		err := client.Ls(ip, port)
		if err != nil {
			log.Println("error", err.Error())
			os.Exit(1)
		}
	case "prune":
		pruneCmd.Parse(os.Args[2:])
		if pruneCmd.NArg() != 0 {
			pruneCmd.Usage()
			os.Exit(0)
		}
		err := client.Prune(ip, port)
		if err != nil {
			log.Println("error", err.Error())
			os.Exit(1)
		}
	case "rm":
		rmCmd.Parse(os.Args[2:])
		if rmCmd.NArg() < 1 {
			rmCmd.Usage()
			os.Exit(0)
		}
		err := client.Rm(ip, port, rmCmd.Arg(0))
		if err != nil {
			log.Println("error", err.Error())
			os.Exit(1)
		}
	default:
		flag.Usage()
		os.Exit(0)
	}
}
