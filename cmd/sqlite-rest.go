package main

import (
	"errors"
	"flag"
	"log"
	"net/http"
	"os"

	"github.com/jonamat/sqlite-rest/pkg/controllers"
	"github.com/julienschmidt/httprouter"
)

const (
	VERSION         = "1.0.0"
	DEFAULT_PORT    = "8080"
	DEFAULT_DB_PATH = "./data.sqlite"
)

var help = flag.Bool("help", false, "Show help")
var port = flag.String("p", DEFAULT_PORT, "Port to listen on")
var dbPath = flag.String("f", DEFAULT_DB_PATH, "Path to sqlite database file")

func main() {
	flag.Parse()

	if *help {
		flag.Usage()
		os.Exit(0)
	}

	_, err := os.Stat(*dbPath)
	if err != nil {
		// Create db if not exits
		if errors.Is(err, os.ErrNotExist) {
			log.Printf("Database not found. Creating new one in %s\n", *dbPath)
			_, err := os.Create(*dbPath)
			if err != nil {
				log.Fatal("Error creating sqlite file. " + err.Error())
			}
		} else {
			log.Fatal("Error reading sqlite file. " + err.Error())
		}
	}
	log.Printf("Using database in %s\n", *dbPath)

	router := httprouter.New()

	router.GET("/:table", controllers.GetAll(*dbPath))
	router.GET("/:table/:id", controllers.Get(*dbPath))
	router.POST("/:table", controllers.Create(*dbPath))
	router.PATCH("/:table/:id", controllers.Update(*dbPath))
	// router.PUT("/:table/:id", controllers.Update(*dbPath))
	router.DELETE("/:table/:id", controllers.Delete(*dbPath))

	router.OPTIONS("/__/exec", controllers.Exec(*dbPath))

	log.Println("Listening on port " + *port)
	log.Fatal(http.ListenAndServe(":"+*port, router))
}
