package main

import (
	"fmt"
	"os"

	"github.com/kumahq/kuma-counter-demo/pkg/api"
)

//go:generate go run .
func main() {
	file, err := os.Create("../../ERRORS.md")
	if err != nil {
		fmt.Println("Error creating file:", err)
		return
	}
	defer file.Close()

	_, _ = file.WriteString("# Error Codes\n\n")
	_, _ = file.WriteString("This document provides a list of error codes used in the application.\n")

	for _, err := range api.ErrorTypes {
		_, _ = file.WriteString(fmt.Sprintf(`
## %s

%s
`, err.Key, err.Description))
	}

	fmt.Println("README file generated: ERRORS.md")
}
