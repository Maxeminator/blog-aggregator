package main

import (
	"context"
	"fmt"

	"github.com/Maxeminator/blog-aggregator/internal/database"
)

func middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error {
	return func(s *state, cmd command) error {

		if s.cfg.CurrentUserName == "" {
			return fmt.Errorf("no user logged in")
		}
		user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
		if err != nil {
			return fmt.Errorf("can't find user %w", err)
		}
		return handler(s, cmd, user)
	}
}
