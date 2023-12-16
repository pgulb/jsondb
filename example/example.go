package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/pgulb/jsondb/db"
	"github.com/pgulb/jsondb/structures"
)

// handling response
// jsondb responds with slice of log lines
// if StatusOk is false, then last line will be error
func handleOutput(c chan structures.Response) {
	select {
	case resp := <-c:
		mlen := len(resp.Message)
		for k, v := range resp.Message {
			if k != mlen-1 {
				log.Println(v)
			}
		}
		if resp.StatusOk {
			log.Println(resp.Message[mlen-1])
		} else {
			log.Fatal(resp.Message[mlen-1])
		}
	case <-time.After(time.Second * 60):
		log.Fatal("no response from jsondb")
	}
}

func main() {
	// channels to communicate with jsondb
	input := make(chan structures.Request)
	output := make(chan structures.Response)
	initialOutput := make(chan structures.Response)

	// set path to config with (--cfg=[[file.json]])
	cmdArgs := os.Args[1:]

	log.Println("starting jsondb goroutine")
	go db.Listen(cmdArgs, input, output, initialOutput)

	// initialOutput occurs once
	// after reading config arguments, json files etc
	handleOutput(initialOutput)

	// possible operations are:
	// set - sets key to value in specific keyFamily
	// get - gets specific key from keyFamily
	// list - lists keyFamilys
	// listkeys - lists keys in specific keyFamily

	// each keyFamily is represented by [[keyFamily]].json file
	// json files are stored in dir specified in config

	v := rand.Intn(99999)
	log.Printf("sending request for set tesciwo.1 to qwe-%v\n", v)
	input <- structures.Request{
		KeyFamily: "tesciwo",
		Key:       "1",
		Value:     fmt.Sprintf("qwe-%v", v),
		Action:    "set",
	}
	handleOutput(output)

	v = rand.Intn(99999)
	log.Printf("sending request for set tesciwo.2 to abc-%v\n", v)
	input <- structures.Request{
		KeyFamily: "tesciwo",
		Key:       "2",
		Value:     fmt.Sprintf("abc-%v", v),
		Action:    "set",
	}
	handleOutput(output)

	log.Println("sending request for get tesciwo.1")
	input <- structures.Request{
		KeyFamily: "tesciwo",
		Key:       "1",
		Action:    "get",
	}
	handleOutput(output)

	log.Println("sending request for list")
	input <- structures.Request{
		Action: "list",
	}
	handleOutput(output)

	log.Println("sending request for listkeys in tesciwo")
	input <- structures.Request{
		KeyFamily: "tesciwo",
		Action:    "listkeys",
	}
	handleOutput(output)

	log.Println("sending request and timeouting")
	input <- structures.Request{
		KeyFamily: "tesciwo",
		Action:    "listkeys",
	}
	time.Sleep(time.Second * 100)

	log.Println("then starting new request")
	input <- structures.Request{
		KeyFamily: "tesciwo",
		Action:    "listkeys",
	}
	handleOutput(output)

	log.Println("closing goroutine")
	input <- structures.Request{
		Action: "quit",
	}
	// One can use built-in function to handle responses
	resp, err := db.HandleOutput(output, 60)
	for _, v := range resp {
		fmt.Println(v)
	}
	if err != nil {
		// last line of responses will be in err
		// if something goes wrong
		log.Fatal(err)
	}
}
