package main

import (
	"fmt"
	"os"

	"github.com/go-nacelle/nacelle"
	"github.com/go-nacelle/pgutil"
	"github.com/rubenv/sql-migrate"
)

func main() {
	db, err := pgutil.Dial(os.Getenv("IJCI_DATABASE_URL"), nacelle.NewNilLogger())
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}

	migrations := &migrate.FileMigrationSource{
		Dir: "/migrations",
	}

	n, err := migrate.Exec(db.DB.DB, "postgres", migrations, migrate.Up)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}

	fmt.Printf("applied %d migrations\n", n)
}
