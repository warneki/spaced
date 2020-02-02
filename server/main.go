package main

import (
    "log"
    "net/http"
)

func main() {
    log.Println("Now server is running on port 8000")
    _ = http.ListenAndServe(":8000", Router())
}
