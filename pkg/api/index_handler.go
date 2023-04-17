package api

import (
	"fmt"
	"net/http"
)

func (s *Server) indexHandler(w http.ResponseWriter, r *http.Request) {

	fmt.Fprintf(w, "<html><head></head><body>")
	fmt.Fprintf(w, "<ul>")
	for _, message := range s.circularBuffer.GetContents() {
		fmt.Fprintf(w, "<li>%v</li>", message)
	}
	fmt.Fprintf(w, "</ul>")
	fmt.Fprintf(w, "</body>")
}
