package utils

import (
	"fmt"

	"baselix/internal/config"
)

// Debug prints only if DEBUG=true
// endline: adds --- line after output
func Debug(msg any, endline bool) {
	if !config.Cfg.Debug {
		return
	}

	fmt.Println(msg)

	if endline {
		fmt.Println("--------------------------------")
	}
}
