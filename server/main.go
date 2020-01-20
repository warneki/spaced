package main

import (
	"log"
	"net/http"
)

func main() {
	log.Println("Now server is running on port 3000")
	_ = http.ListenAndServe(":3000", Router())
}
