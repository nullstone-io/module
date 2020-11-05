package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/nullstone-io/module/config"
)

func main() {
	files, err := config.ReadDir(".")
	if err != nil {
		log.Fatal(err)
	}
	cfg, err := config.Parse(files)
	if err != nil {
		log.Fatal(err)
	}

	manifest := cfg.ToManifest()
	raw, _ := json.Marshal(manifest)
	fmt.Println(string(raw))
}
