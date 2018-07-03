package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/BillyPurvis/boommessaging-go/database"
	"github.com/BillyPurvis/boommessaging-go/ldaphandler"
	"github.com/BillyPurvis/boommessaging-go/middleware"
	_ "github.com/joho/godotenv/autoload"
	"github.com/julienschmidt/httprouter"

	_ "github.com/go-sql-driver/mysql"
)

func main() {

	// Make DB Connection
	var err error
	database.DBCon, err = sql.Open("mysql", "root:root@/boom")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Starting Server on port %v:%v\n", os.Getenv("APP_URL"), os.Getenv("APP_PORT"))

	// Create Go Server
	router := httprouter.New()

	router.POST("/", middleware.AuthenticateWare(ldaphandler.GetAttributes))
	router.POST("/ldap", middleware.AuthenticateWare(ldaphandler.GetAttributes))

	log.Fatal(http.ListenAndServe(":4000", middleware.SetJSONHeader(router)))
}
