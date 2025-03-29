package handlers

import (
	"net/http"
	"fmt"
)

func HelloWorldHandler(w http.ResponseWriter, r *http.Request){
	fmt.Fprintf(w, "Hello World!")
}