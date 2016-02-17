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
	router.HandleFunc("/list", List)
	log.Fatal(http.ListenAndServe(":8080", router))
}

func List(w http.ResponseWriter, r *http.Request) {
	cal := NewCalendarService()
	rooms := cal.GetAvaliableRooms()
	fmt.Fprintf(w, "%q currently avaliable", html.EscapeString(rooms))
}
