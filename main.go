package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html"
	"log"
	"net/http"
)

func main() {
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/", Index)
	router.HandleFunc("/book", Book)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func Index(w http.ResponseWriter, r *http.Request) {
	log.Printf("%d", r)
	rooms := NewCalendarService().GetAvaliableRooms()
	fmt.Fprintf(w, "Hello, %q rooms avaliable", html.EscapeString(rooms))
}

func Book(w http.ResponseWriter, r *http.Request) {
	res := NewCalendarService().QuickBook("movableink.com_j7it2087rr0psr85m6nttp313k%40group.calendar.google.com")
	fmt.Fprintf(w, "Booked", res)
}
