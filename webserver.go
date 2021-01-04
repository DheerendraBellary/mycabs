package main

import (
	"fmt"
	"mycabs/mycabsservice"
	"net/http"
	"os"
)

func port() string {
	port := os.Getenv("MYCABS_WEBSERVER_PORT")
	if port == "" {
		return ":8080"
	}
	return ":" + port
}

func main() {
	fmt.Println("MyCabs Webserver running....")
	fmt.Printf("Port: %v", port())

	http.HandleFunc("/api/OnboardCity", mycabsservice.OnboardCityHandler)
	http.HandleFunc("/api/RegisterCab", mycabsservice.RegisterCabHandler)
	http.HandleFunc("/api/BookCab", mycabsservice.BookCabHandler)
	http.HandleFunc("/api/EndTrip", mycabsservice.EndTripHandler)
	http.HandleFunc("/api/DeActivateCab", mycabsservice.DeActivateCabHandler)
	http.HandleFunc("/api/ActivateCab", mycabsservice.ActivateCabHandler)
	http.HandleFunc("/api/ChangeCity", mycabsservice.ChangeCityHandler)

	http.ListenAndServe(port(), nil)
}
