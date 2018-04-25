package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"
)

var tickets *TicketStore
var renderer *render.Render
var port string
var host string

func init() {
	flag.StringVar(&port, "port", "3000", "Port to run the web server on.")
	flag.StringVar(&host, "host", "localhost", "Hostname/IP to run the web server on.")
	flag.Parse()

	tickets = NewTicketStore()
	renderer = render.New()
}

func checkError(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	r := mux.NewRouter()
	r.HandleFunc("/tickets", CreateTicketHandler).Methods("POST")
	r.HandleFunc("/tickets/{id}", ShowTicketTotalHandler).Methods("GET")
	r.HandleFunc("/payments/{ticketid}", PayTicketHandler).Methods("POST")

	n := negroni.Classic()
	n.UseHandler(r)

	listenAddress := fmt.Sprintf("%s:%s", host, port)

	h := &http.Server{Addr: listenAddress, Handler: n}

	go func() {

		log.Println("Server running @ " + listenAddress)

		if err := h.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	<-stop

	log.Println("\nShutting down...")

	tickets.db.Close()

	h.Shutdown(context.Background())
}

// CreateTicketHandler will print a new ticket for a user
func CreateTicketHandler(w http.ResponseWriter, r *http.Request) {
	newTicket := tickets.Create()

	renderer.JSON(w, http.StatusOK, map[string]int64{"ticket_number": newTicket.ID})
}

// ShowTicketTotalHandler will show the user how much they owe
func ShowTicketTotalHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)

	checkError(err)

	ticket, err := tickets.Find(id)

	if err != nil {
		renderer.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	renderer.JSON(w, http.StatusOK, ticket.ShowOwing())
}

// PayTicketHandler will allow a user to pay a ticket by accepting a credit card and ticket id
func PayTicketHandler(w http.ResponseWriter, r *http.Request) {

	vars := mux.Vars(r)

	id, err := strconv.ParseInt(vars["ticketid"], 10, 64)
	checkError(err)

	if r.Body == nil {
		http.Error(w, "No request body sent.", 400)
		return
	}

	cardInfo := struct {
		CreditCard string `json:"credit_card"`
	}{}

	err = json.NewDecoder(r.Body).Decode(&cardInfo)

	if err != nil {
		http.Error(w, err.Error(), 400)
		return
	}

	ticket, err := tickets.Pay(id, cardInfo.CreditCard)

	if err != nil {
		renderer.JSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
		return
	}

	renderer.JSON(w, http.StatusOK, ticket)
}
