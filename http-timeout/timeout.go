package main

import (
	"fmt"
	"net/http"
	"time"

	"flag"

	"github.com/julienschmidt/httprouter"
)

func main() {
	timeout := flag.Int("timeout", 10, "Timeout in seconds before response is resnt")
	flag.Parse()

	router := httprouter.New()
	router.GET("/*ignore", func(res http.ResponseWriter, req *http.Request, _ httprouter.Params) {
		fmt.Printf("Delaying %s for %d second(s)", req.URL.String(), *timeout)
		time.Sleep(time.Duration(*timeout) * time.Second)
		res.Write([]byte("ok"))
	})
	fmt.Println("Listening on :8080")
	http.ListenAndServe(":8080", router)
}
