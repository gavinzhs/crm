package main

import (
	"flag"
	"log"
	"net/http"
)

const (
	SALT   = "Pcmmr]g[*Z'v"
	DB_CON = "crm:crm@:27017/crm?maxPoolSize=50"
)

func main() {
	print("start crm\n")

	addr := flag.String("p", ":3000", "address where the server listen on")
	war := flag.String("war", "./public", "directory of war files")
	flag.Parse()

	//connect mongodb
	log.Println("connect mongo")
	session := connect(DB_CON)

	//init seq
	initSeq(session)

	log.Println("ensure index")
	ensureIndex(session)

	initData(session)

	//add static file
	mount(session, *war)

	log.Fatal(http.ListenAndServe(*addr, nil))
}
