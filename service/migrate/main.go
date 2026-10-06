package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	salemongo "github.com/MamangRust/monolith-payment-gateway-pkg/database/mongo"
	"github.com/MamangRust/monolith-payment-gateway-pkg/database/seeder"
	"github.com/MamangRust/monolith-payment-gateway-pkg/dotenv"
	"github.com/MamangRust/monolith-payment-gateway-pkg/hash"
	"github.com/MamangRust/monolith-payment-gateway-pkg/logger"
)

var flags = flag.NewFlagSet("migrate", flag.ExitOnError)

// main bootstraps the MongoDB collections and indexes used by the
// payment-gateway services. Schema-less storage means there are no SQL
// migrations anymore: "up" applies the index definitions, "status"
// reports the collections that currently exist, and "seed" runs the
// data seeder.
func main() {
	flags.Usage = usage
	if err := flags.Parse(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v\n", err)
		os.Exit(1)
	}

	args := flags.Args()
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" {
		flags.Usage()
		return
	}

	command := args[0]

	if err := dotenv.Viper(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	l, err := logger.NewLogger("migrate", nil)
	if err != nil {
		log.Fatalf("Error creating logger: %v", err)
	}

	ctx := context.Background()

	client, err := salemongo.NewMongoClient(l)
	if err != nil {
		log.Fatalf("Error connecting to MongoDB: %v", err)
	}
	defer func() {
		if err := client.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting MongoDB client: %v", err)
		}
	}()

	db := salemongo.NewDatabase(client)

	switch command {
	case "up":
		if err := salemongo.CreateIndexes(ctx, db); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		fmt.Println("MongoDB collections and indexes are up to date")

	case "status":
		names, err := db.ListCollectionNames(ctx, map[string]any{})
		if err != nil {
			log.Fatalf("Failed to list collections: %v", err)
		}
		fmt.Printf("MongoDB database %q collections:\n", db.Name())
		for _, name := range names {
			fmt.Printf("    %s\n", name)
		}

	case "seed":
		s := seeder.NewSeeder(seeder.Deps{
			DB:     db,
			Hash:   hash.NewHashingPassword(),
			Ctx:    ctx,
			Logger: l,
		})
		if err := s.Run(); err != nil {
			log.Fatalf("Seeding failed: %v", err)
		}
		fmt.Println("MongoDB seeding complete")

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", command)
		flags.Usage()
		os.Exit(1)
	}

}

func usage() {
	fmt.Println(usagePrefix)
	flags.PrintDefaults()
	fmt.Println(usageCommands)
}

var (
	usagePrefix = `Usage: migrate COMMAND
Examples:
    migrate up
`

	usageCommands = `
Commands:
    up                   Apply MongoDB collections and indexes
    status               List the collections in the target database
    seed                 Seed sample data into MongoDB`
)
