package main

import (
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"os"

	"github.com/julienschmidt/httprouter"
)

// func createKeyValuePairs(m map[string][]string) string {
// 	b := new(bytes.Buffer)
// 	for key, value := range m {
// 		fmt.Fprintf(b, "%s=\"%s\"", key, value[0])
// 	}
// 	return b.String()
// }

func hash(s string) uint32 {
	h := fnv.New32a()
	h.Write([]byte(s))
	return h.Sum32()
}

func handler(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	fmt.Fprint(w, "Welcome!\n")
	fmt.Fprintf(w, "Hi there, I love  %s!", r.URL.Path[1:])
	fmt.Fprintf(w, "hello, %s!\n", ps.ByName("name"))

	user := ps.ByName("user")
	resource_modifiers := ps.ByName("resource") + ps.ByName("modifiers")
	resource_hash := fmt.Sprint(hash(resource_modifiers))

	log.Println(user)
	log.Println(resource_hash)
	log.Println("Check if resource exists in cache for given parameters")
}

func main() {
	port := os.Getenv("PORT")
	log.Println(port)
	if port == "" {
		log.Fatal("$PORT must be set")
	}
	router := httprouter.New()
	router.GET("/user/:user/resource/:resource/modifiers/:modifiers", handler)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
