package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

func main() {
	// Find all directories matching the pattern
	files, err := os.ReadDir("kustomize/overlays")
	if err != nil {
		panic(fmt.Errorf("error reading current directory %w", err))
	}

	maxNum := 0
	allNames := []string{}
	for _, file := range files {
		if file.IsDir() {
			p := strings.Split(file.Name(), "-")
			if len(p) >= 2 {
				num, err := strconv.Atoi(p[0])
				if err == nil {
					if num > maxNum {
						maxNum = num
					}
					allNames = append(allNames, file.Name())
				}
			}
		}
	}
	slices.Sort(allNames)
	nextNumber := maxNum + 1

	// Prompt user for the name
	fmt.Printf("Enter a name for the new entry for demo %d: ", nextNumber)
	reader := bufio.NewReader(os.Stdin)
	name, err := reader.ReadString('\n')
	if err != nil {
		panic(fmt.Errorf("error reading input %w", err))
	}
	name = strings.TrimSpace(name)
	if m, err := regexp.MatchString(`^[A-Za-z0-9-]+$`, name); !m || err != nil {
		panic("Error: Name must have only letters, numbers or `-`!")
	}
	outFiles := map[string]string{}

	fmt.Print("Enter a one line description:")
	desc, err := reader.ReadString('\n')
	if err != nil {
		panic(fmt.Errorf("error reading input %w", err))
	}
	outFiles["README.md"] = fmt.Sprintf("# %s\n", strings.TrimSpace(desc))

	fmt.Printf("Select the existing demo that this new entry should be based on:\n%s\n---:61\n", strings.Join(allNames, "\n"))
	idxStr, err := reader.ReadString('\n')
	if err != nil {
		panic(fmt.Errorf("error reading input %w", err))
	}
	idx, err := strconv.Atoi(strings.Split(strings.TrimSpace(idxStr), "-")[0])
	if err != nil || idx < 0 || idx >= len(allNames) {
		panic(fmt.Sprintf("Invalid selection %q", idxStr))
	}
	outFiles["kustomization.yaml"] = fmt.Sprintf(`---
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
resources:
  - ../%s
`, allNames[idx])

	// Create the folder and the files
	newFolder := fmt.Sprintf("kustomize/overlays/%03d-%s", nextNumber, name)
	if err := os.Mkdir(newFolder, 0755); err != nil {
		panic(fmt.Errorf("error creating folder %w", err))
	}
	for f := range outFiles {
		createFile(newFolder, f, outFiles[f])
	}
	fmt.Println("Created folder", newFolder)
}

func createFile(dir, name, content string) {
	fmt.Printf("Creating %q\n", name)
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		panic(fmt.Errorf("error creating %q %w", path, err))
	}
}
