package main

import (
	"fmt"
	"log"

	envyaml "go.tmthrgd.dev/env-yaml"
)

func main() {
	env, err := envyaml.ShellEscaped()
	if err != nil {
		log.Fatal(err)
	}

	for _, kv := range env {
		fmt.Println(kv)
	}
}
