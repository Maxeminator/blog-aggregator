package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/Maxeminator/blog-aggregator/internal/config"
	"github.com/Maxeminator/blog-aggregator/internal/database"
	"github.com/google/uuid"
)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

type command struct {
	name string
	args []string
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("username required")
	}
	name := cmd.args[0]
	_, err := s.db.GetUser(context.Background(), name)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("user does not exist")
	}
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	err = s.cfg.SetUser(cmd.args[0])
	if err != nil {
		return fmt.Errorf("can't set the username: %w", err)
	}
	fmt.Printf("username set to: %s\n", cmd.args[0])
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("username required")
	}
	name := cmd.args[0]
	now := time.Now()
	_, err := s.db.GetUser(context.Background(), name)
	if err == nil {

		return fmt.Errorf("user already exists")
	}
	if !errors.Is(err, sql.ErrNoRows) {

		return fmt.Errorf("failed to check user: %w", err)
	}

	user, err := s.db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: now,
		UpdatedAt: now,
		Name:      name,
	})
	if err != nil {
		return fmt.Errorf("create user: %w", err)
	}

	err = s.cfg.SetUser(name)
	if err != nil {
		return fmt.Errorf("failed to set current user: %w", err)
	}

	fmt.Printf("user created: id=%s, name=%s\n", user.ID, user.Name)
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
