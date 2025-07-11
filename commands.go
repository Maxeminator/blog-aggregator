package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
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

func handlerReset(s *state, cmd command, user database.User) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to reset database %w", err)
	}
	fmt.Println("reset complete")
	return nil
}

func handlerUsers(s *state, cmd command, user database.User) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("failed to check users %w", err)

	}
	for _, user := range users {
		name := user.Name
		if name == s.cfg.CurrentUserName {
			fmt.Printf("* %s (current)\n", name)
		} else {
			fmt.Printf("* %s\n", name)
		}
	}
	return nil
}

func handlerAgg(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("usage: agg <duration>")
	}

	duration, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return fmt.Errorf("invalid duration: %w", err)
	}

	fmt.Printf("Collecting feeds every %s\n", duration)

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		err := handlerScrapeFeeds(s, cmd, user)
		if err != nil {
			fmt.Printf("Error scraping: %v\n", err)
		}
		<-ticker.C
	}
}

func handlerAddfeed(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 2 {
		return fmt.Errorf("usage: addfeed <name> <url>")
	}

	name := cmd.args[0]
	url := cmd.args[1]

	if s.cfg.CurrentUserName == "" {
		return fmt.Errorf("no user logged in")
	}
	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), params)

	if err != nil {
		return fmt.Errorf("can't create feed database: %w", err)
	}

	fmt.Printf("%+v\n", feed)

	follows := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	follow, err := s.db.CreateFeedFollow(context.Background(), follows)
	if err != nil {
		return fmt.Errorf("could not follow feed: %w", err)
	}
	fmt.Printf("Feed %q successfully added and followed by %q\n", follow.FeedName, follow.UserName)
	return nil

}

func handlerFeeds(s *state, cmd command, user database.User) error {
	feeds, err := s.db.ListFeedsWithUsers(context.Background())
	if err != nil {
		return fmt.Errorf("can't read the feed %w", err)
	}
	for _, f := range feeds {
		fmt.Printf("Name: %s\nURL: %s\nUser: %s\n\n", f.Name, f.Url, f.Name_2)
	}
	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("usage: follow <url>")
	}

	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("feed not found: %w", err)
	}
	follows := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feed.ID,
	}

	follow, err := s.db.CreateFeedFollow(context.Background(), follows)
	if err != nil {
		return fmt.Errorf("could not follow feed: %w", err)
	}
	fmt.Printf("Following to %q as %q\n", follow.FeedName, follow.UserName)
	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	follows, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return fmt.Errorf("failed to get subscriptions: %w", err)
	}

	if len(follows) == 0 {
		fmt.Println("You are not following any feeds.")
		return nil
	}

	for _, f := range follows {
		fmt.Printf("Name: %s\n", f.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.args) < 1 {
		return fmt.Errorf("usage: unfollow <url>")
	}
	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("can't find the feed %w", err)
	}
	unfollow := database.UnfollowUserParams{
		UserID: user.ID,
		FeedID: feed.ID,
	}

	err = s.db.UnfollowUser(context.Background(), unfollow)
	if err != nil {
		return fmt.Errorf("can't unfollow %w", err)
	}
	fmt.Printf("Unfollowed from \"%s\"\n", feed.Name)
	return nil
}

func parseTime(dateStr string) (time.Time, error) {
	formats := []string{
		time.RFC1123,
		time.RFC1123Z,
		time.RFC3339,
	}
	var t time.Time
	var err error
	for _, layout := range formats {
		t, err = time.Parse(layout, dateStr)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("oculd not parse date %q: %v", dateStr, err)
}

func handlerScrapeFeeds(s *state, cmd command, user database.User) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("can't find feed to fetch %w", err)
	}
	err = s.db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{Time: time.Now(), Valid: true},
		ID:            feed.ID,
	})
	if err != nil {
		return fmt.Errorf("failed to mark feed as fetched: %w", err)
	}

	rssFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return fmt.Errorf("failed to fetch RSS feed: %w", err)
	}

	fmt.Printf("Fetched %d posts from feed: %s\n", len(rssFeed.Channel.Item), feed.Name)
	for _, item := range rssFeed.Channel.Item {
		published, err := parseTime(item.PubDate)
		if err != nil {
			log.Printf("can't parse date %q: %v", item.PubDate, err)
			continue
		}
		_, err = s.db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: published,
			FeedID:      feed.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			log.Printf("failed to insert post: %v", err)
		}
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.args) >= 1 {
		parsedLimit, err := strconv.Atoi(cmd.args[0])
		if err != nil {
			return fmt.Errorf("invalid limit: %w", err)
		}
		limit = parsedLimit
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("failed to get posts: %w", err)
	}

	if len(posts) == 0 {
		fmt.Println("no posts found.")
		return nil
	}

	for _, post := range posts {
		fmt.Printf("Title: %s\nUrl: %s\nPublished: %s\nFeed: %s\n\n", post.Title, post.Url, post.PublishedAt, post.FeedID)
	}

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
