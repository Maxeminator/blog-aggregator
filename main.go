package main

import (
	"fmt"
	"log"
	"os"

	"github.com/Maxeminator/blog-aggregator/internal/config"
)

func main() {
	cfg := config.Config{}
	err := config.Read(&cfg)
	if err != nil {
		log.Fatal(err)
	}
	st := &state{config: &cfg}

	cmds := &commands{handlers: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	args := os.Args
	if len(args) < 2 {
		fmt.Println("not enough arguments")
		os.Exit(1)
	}
	cmd := command{
		name: args[1],
		args: args[2:],
	}
	err = cmds.run(st, cmd)
	if err != nil {
		fmt.Printf("error: %v\n", err)
		os.Exit(1)
	}
}
