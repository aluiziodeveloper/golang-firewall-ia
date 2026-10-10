package main

import (
	"fmt"
	"io"
	"log"
	"os"
)

const startupMessage = "Guarita: fundação inicial; o proxy MCP ainda não está implementado."

func main() {
	if err := run(os.Stderr); err != nil {
		log.Printf("não foi possível iniciar: %v", err)
		os.Exit(1)
	}
}

// #region cmd-guarita-startup
func run(stderr io.Writer) error {
	_, err := fmt.Fprintln(stderr, startupMessage)
	return err
}

// #endregion
