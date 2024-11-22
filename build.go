// +build ignore

package main

import (
	"fmt"
	"os"
	"os/exec"
)

func main() {
	cmd := exec.Command("go", "build", "-tags", "dev", "-o", "awesome-cv.exe", "cmd/main.go")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	fmt.Println("🔨 Bygger awesome-cv.exe...")
	if err := cmd.Run(); err != nil {
		fmt.Printf("❌ Byggfel: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✅ Bygget klart!")
}
