package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Maxeminator/blog-aggregator/internal/config"
	"github.com/Maxeminator/blog-aggregator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	dbURL := "postgres://postgres:postgres@localhost:5432/gator?sslmode=disable"
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("can't open db: %v", err)
	}
	dbQueries := database.New(db)

	cfg := config.Config{}
	err = config.Read(&cfg)
	if err != nil {
		log.Fatal(err)
	}

	st := &state{
		db:  dbQueries,
		cfg: &cfg,
	}

	cmds := &commands{handlers: make(map[string]func(*state, command) error)}
	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)

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
