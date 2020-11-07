package contactbook

import "context"

type Book interface {
	AddContact(ctx context.Context, contact *Contact) error
	UpdateContact(ctx context.Context, id int, contact *Contact) error
	DeleteContact(ctx context.Context, id int) error
	ListContacts(ctx context.Context, cb func(contacts []*Contact) error) error
}

type Contact struct {
	Id               int      `json:"id"`
	FirstName        string   `json:"first_name"`
	LastName         string   `json:"last_name"`
	TelephoneNumbers []string `json:"telephone_numbers"`
}
