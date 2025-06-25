// Package main is the entry point for the CrowNet application.
// It initializes the configuration, sets up the command-line interface (CLI)
// orchestrator, and runs the simulation based on the provided arguments.
package main

import (
	"crownet/cmd"
	// "crownet/cli" // A lógica do orchestrator será chamada pelos comandos do Cobra
	// "crownet/config" // A configuração será gerenciada dentro dos comandos Cobra
	// "fmt"
	// "log"
	// "os"
)

func main() {
	// REFACTOR-CLI-001: A lógica de parsing de flags e execução de modo
	// é agora gerenciada pelo pacote cmd (usando Cobra).
	cmd.Execute()
}
