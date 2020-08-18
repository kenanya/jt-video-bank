package main

import (
	"fmt"
	"os"

	"github.com/kenanya/jt-video-bank/pkg/cmd" 
	// "github.com/kenanya/jt-video-bank/pkg/api/v1"
)

func main() {
	if err := cmd.RunServer(); err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
}
