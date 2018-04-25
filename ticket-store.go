package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

func itob(v int64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, uint64(v))
	return b
}

// TicketStore is where the tickets live, get you some
type TicketStore struct {
	db     *bolt.DB
	bucket []byte
}

// NewTicketStore gets you a new ticket store
func NewTicketStore() *TicketStore {

	database, err := bolt.Open("parkypark.db", 0600, nil)

	checkError(err)

	bucketName := []byte("Tickets")

	database.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists(bucketName)

		checkError(err)
		return nil
	})

	return &TicketStore{
		db:     database,
		bucket: bucketName,
	}
}

// Create a new ticket
func (ts *TicketStore) Create() *Ticket {

	now := time.Now()

	newTicket := &Ticket{
		TimeIn: now.Add(time.Duration(-350) * time.Minute),
	}

	err := ts.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ts.bucket)

		id, _ := b.NextSequence()
		newTicket.ID = int64(id)

		buf, err := json.Marshal(newTicket)
		checkError(err)

		return b.Put(itob(newTicket.ID), buf)
	})

	checkError(err)

	return newTicket
}

// Find will get a ticket from the database
func (ts *TicketStore) Find(ticketID int64) (*Ticket, error) {
	ticket := &Ticket{}

	err := ts.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(ts.bucket)
		v := b.Get(itob(ticketID))

		json.Unmarshal(v, &ticket)
		return nil
	})

	checkError(err)

	if ticket.ID == 0 {
		return nil, fmt.Errorf("Cannot find A ticket with ID: %v", ticketID)
	}

	return ticket, nil
}

// Save will update the ticket in the store
func (ts *TicketStore) Save(ticket *Ticket) error {
	return ts.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(ts.bucket)

		buf, err := json.Marshal(ticket)
		checkError(err)

		return b.Put(itob(ticket.ID), buf)

	})
}

// Pay the ticket off
func (ts *TicketStore) Pay(ticketID int64, ccNumber string) (*Ticket, error) {
	// validate cc number, yadda yadda yadda
	t, err := tickets.Find(ticketID)

	checkError(err)

	t.Paid = int64(t.CalculateOwing())
	t.Card = ccNumber

	return t, ts.Save(t)
}
