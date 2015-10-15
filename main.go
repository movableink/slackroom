package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"html"
	"log"
	"net/http"
)

type Main struct {
	cal *Calendar
}

func main() {
	x := NewStruct()
	router := mux.NewRouter().StrictSlash(true)
	router.HandleFunc("/list", x.Index)
	// router.HandleFunc("/book", Book)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func NewStruct() *Main {
	cal := NewCalendarService()
	return &Main{cal: cal}
}

func (main *Main) Index(w http.ResponseWriter, r *http.Request) {
	log.Printf("%d", r)
	rooms := main.cal.GetAvaliableRooms()
	fmt.Fprintf(w, "%q currently avaliable", html.EscapeString(rooms))
}

// func Book(w http.ResponseWriter, r *http.Request) {
// 	res := NewCalendarService().QuickBook("movableink.com_j7it2087rr0psr85m6nttp313k%40group.calendar.google.com")
// 	fmt.Fprintf(w, "Booked", res)
// }
