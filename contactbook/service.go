package contactbook

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"

	"github.com/ykhrustalev/contacts-book/storage"
)

var (
	ErrContactExists  = errors.New("contact exists")
	ErrContactMissing = errors.New("contact missing")
	ErrContactInvalid = errors.New("contact invalid")
)

type BookImpl struct {
	store storage.Store

	topId            int
	contactsById     map[int]*Contact
	contactsByUnique map[string]*Contact

	mu sync.RWMutex
}

func New(store storage.Store) (*BookImpl, error) {
	book := &BookImpl{
		store:            store,
		contactsById:     map[int]*Contact{},
		contactsByUnique: map[string]*Contact{},
	}
	return book, book.load()
}

func uniqueKey(contact *Contact) string {
	return fmt.Sprintf("%s:%s", contact.FirstName, contact.LastName)
}

func validate(contact *Contact) error {
	if contact.FirstName == "" || contact.LastName == "" {
		return fmt.Errorf("%w, one of first or last name should not be blank", ErrContactInvalid)
	}

	return nil
}

func (book *BookImpl) AddContact(ctx context.Context, contact *Contact) error {
	book.mu.Lock()
	defer book.mu.Unlock()

	if err := validate(contact); err != nil {
		return err
	}

	uniqKey := uniqueKey(contact)
	_, ok := book.contactsByUnique[uniqKey]
	if ok {
		return fmt.Errorf("%w for %s %s", ErrContactExists, contact.FirstName, contact.LastName)
	}

	book.topId++
	contact.Id = book.topId
	book.contactsById[book.topId] = contact
	book.contactsByUnique[uniqKey] = contact

	return book.save()
}

func (book *BookImpl) UpdateContact(ctx context.Context, id int, contact *Contact) error {
	book.mu.Lock()
	defer book.mu.Unlock()

	existing, ok := book.contactsById[id]
	if !ok {
		return fmt.Errorf("%w for %d", ErrContactMissing, id)
	}

	if err := validate(contact); err != nil {
		return err
	}

	existing.FirstName = contact.FirstName
	existing.LastName = contact.LastName
	existing.TelephoneNumbers = contact.TelephoneNumbers

	return book.save()
}

func (book *BookImpl) DeleteContact(ctx context.Context, id int) error {
	book.mu.Lock()
	defer book.mu.Unlock()

	existing, ok := book.contactsById[id]
	if !ok {
		return fmt.Errorf("%w for %d", ErrContactMissing, id)
	}

	uniqKey := uniqueKey(existing)
	delete(book.contactsById, existing.Id)
	delete(book.contactsByUnique, uniqKey)

	return book.save()
}

func (book *BookImpl) ListContacts(ctx context.Context, cb func(contacts []*Contact) error) error {
	book.mu.RLock()
	defer book.mu.RUnlock()

	// todo: implement callback iterator
	var contacts []*Contact
	for _, item := range book.contactsById {
		contacts = append(contacts, item)
	}

	sort.Slice(contacts, func(i, j int) bool {
		return contacts[i].Id < contacts[j].Id
	})

	return cb(contacts)
}

func (book *BookImpl) load() error {
	b, err := book.store.Read()
	if err != nil {
		return err
	}

	var contacts []*Contact
	if len(bytes.TrimSpace(b)) == 0 {
		// empty
		return nil
	}

	if err := json.Unmarshal(b, &contacts); err != nil {
		return err
	}

	for _, contact := range contacts {
		book.contactsById[contact.Id] = contact
		book.contactsByUnique[uniqueKey(contact)] = contact
		if book.topId < contact.Id {
			book.topId = contact.Id
		}
	}
	return nil
}

func (book *BookImpl) save() error {
	var contacts []*Contact

	for _, item := range book.contactsById {
		contacts = append(contacts, item)
	}

	b, err := json.Marshal(contacts)
	if err != nil {
		return err
	}

	return book.store.Write(b)
}
