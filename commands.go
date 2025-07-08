package main

import (
	"fmt"

	"github.com/Maxeminator/blog-aggregator/internal/config"
)

type state struct {
	config *config.Config
}

type command struct {
	name string
	args []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("username required")
	}
	err := s.config.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("can't set the username: %w", err)
	}
	fmt.Printf("username set to: %s\n", cmd.args[0])
	return nil
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.handlers[name] = f
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command: %s", cmd.name)
	}

	return handler(s, cmd)
}
