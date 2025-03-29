package helloWorld

import (
	"net/http"
	"fmt"
)

func HelloWorldHandler(w http.ResponseWriter, r *http.Request){

	fmt.Println("Testing")
	fmt.Fprintf(w, "Hello World!")
}
