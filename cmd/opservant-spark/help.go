package main

import (
	"fmt"
	"runtime"
)

func printHelp() {
	fmt.Printf("Opservant Spark — security agent\n")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  spark playbook [--config <path>] <file>")
	fmt.Println("  spark playbook [--config <path>] --dir <dir>")
	fmt.Println("  spark version")
	fmt.Println("  spark help")
	fmt.Println()
	fmt.Println("Playbook flags:")
	fmt.Println("  --config <path>       Path to config.yaml (default: config.yaml)")
	fmt.Println("  --dir <dir>  Directory of .json playbooks to execute")
	fmt.Println()
	fmt.Println("Examples:")
	switch runtime.GOOS {
	case "windows":
		fmt.Println(`  .\spark playbook ram-info.json`)
		fmt.Println(`  .\spark playbook --config config.yaml ram-info.json`)
		fmt.Println(`  .\spark playbook --dir .\playbooks`)
	default:
		fmt.Println("  ./spark playbook ram-info.json")
		fmt.Println("  ./spark playbook --config config.yaml ram-info.json")
		fmt.Println("  ./spark playbook --dir ./playbooks")
	}
}
