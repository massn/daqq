package main

import (
	"fmt"
	"log"
	"time"

	"github.com/massn/daqq/circuit"
)

func main() {
	seed := uint32(time.Now().UnixNano())
	width := uint32(8)
	depth := uint32(32)
	randomQC := circuit.MakeRandomQC(seed, width, depth)

	if err := randomQC.State(); err != nil {
		log.Fatalf("Failed to generate quantum state: %v", err)
	}

	fmt.Println("Visualization saved to line.html")
}
