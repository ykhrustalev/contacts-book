package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/urfave/cli/v2"

	"github.com/ykhrustalev/contacts-book/contactbook"
	"github.com/ykhrustalev/contacts-book/storage"
)

var logger *log.Logger

func init() {
	logger = log.New(os.Stderr, "", 0)
	logger.SetPrefix("//")
}

func withInterrupt(ctx context.Context, fn func(ctx context.Context) error) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		wg.Done()
		<-quit
		cancel()
	}()
	wg.Wait()

	return fn(ctx)
}

func main() {
	err := withInterrupt(context.Background(), run)
	if err != nil {
		logger.Fatal(err)
	}
}

const (
	flagId        = "id"
	flagFirstName = "first-name"
	flagLastName  = "last-name"
	flagPhone     = "phone"
)

func clean(v string) string {
	return strings.TrimSpace(v)
}

func loadContact(c *cli.Context) *contactbook.Contact {
	contact := &contactbook.Contact{
		FirstName: clean(c.String(flagFirstName)),
		LastName:  clean(c.String(flagLastName)),
	}

	for _, item := range strings.Split(c.String(flagPhone), ",") {
		item = clean(item)
		if len(item) > 0 {
			contact.TelephoneNumbers = append(contact.TelephoneNumbers, item)
		}
	}

	return contact
}

func run(ctx context.Context) error {
	app := cli.NewApp()
	app.Description = "A simple cli contacts book that stores data locally"
	app.Usage = "Run -help to see available commands"
	app.Commands = cli.Commands{
		&cli.Command{
			Name: "list",
			Action: func(c *cli.Context) error {
				return withFileStorage(ctx, func(storage *storage.FileStore) error {
					return withContactBook(storage, func(book contactbook.Book) error {
						return book.ListContacts(ctx, func(contacts []*contactbook.Contact) error {
							fmt.Println("# contacts")
							if len(contacts) == 0 {
								fmt.Println("(empty)")
							}
							for _, contact := range contacts {
								fmt.Printf("[%d] %s %s\n",
									contact.Id, contact.FirstName, contact.LastName)
								for _, phone := range contact.TelephoneNumbers {
									fmt.Printf(" - %s\n", phone)
								}
							}

							return nil
						})
					})
				})
			},
		},
		&cli.Command{
			Name: "add",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: flagFirstName},
				&cli.StringFlag{Name: flagLastName},
				&cli.StringFlag{Name: flagPhone, Usage: "use comma-separated list for several values"},
			},
			Action: func(c *cli.Context) error {
				return withFileStorage(ctx, func(storage *storage.FileStore) error {
					return withContactBook(storage, func(book contactbook.Book) error {
						contact := loadContact(c)

						err := book.AddContact(ctx, contact)
						if err != nil {
							return err
						}

						fmt.Println("succeed")
						return nil
					})
				})
			},
		},
		&cli.Command{
			Name: "edit",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: flagId},
				&cli.StringFlag{Name: flagFirstName},
				&cli.StringFlag{Name: flagLastName},
				&cli.StringFlag{Name: flagPhone, Usage: "use comma-separated list for several values"},
			},
			Action: func(c *cli.Context) error {
				return withFileStorage(ctx, func(storage *storage.FileStore) error {
					return withContactBook(storage, func(book contactbook.Book) error {
						id := c.Int(flagId)
						contact := loadContact(c)

						err := book.UpdateContact(ctx, id, contact)
						if err != nil {
							return err
						}

						fmt.Println("succeed")
						return nil
					})
				})
			},
		},
		&cli.Command{
			Name: "delete",
			Flags: []cli.Flag{
				&cli.StringFlag{Name: flagId},
			},
			Action: func(c *cli.Context) error {
				return withFileStorage(ctx, func(storage *storage.FileStore) error {
					return withContactBook(storage, func(book contactbook.Book) error {
						id := c.Int(flagId)

						err := book.DeleteContact(ctx, id)
						if err != nil {
							return err
						}

						fmt.Println("succeed")
						return nil
					})
				})
			},
		},
	}
	return app.RunContext(ctx, os.Args)
}

func withContactBook(store storage.Store, cb func(book contactbook.Book) error) error {
	book, err := contactbook.New(store)
	if err != nil {
		return err
	}
	return cb(book)
}

func withFileStorage(ctx context.Context, cb func(storage *storage.FileStore) error) error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dbFile := homeDir + "/.contacts-book-sample-delete-me"

	lockFile := os.TempDir() + "/contacts-book.lock"

	logger.Printf("using storage %s", dbFile)
	logger.Printf("with lock %s", lockFile)

	store, err := storage.NewFileStore(ctx, dbFile, lockFile)
	if err != nil {
		return err
	}

	defer func() {
		if err := store.Close(); err != nil {
			logger.Printf("failed to close storage, %v", err)
		}
	}()

	return cb(store)
}
