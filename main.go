package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
	"time"
)

var harPtr *bool
var outPtr *string
var har *HarLog

func handler(w http.ResponseWriter, r *http.Request) {
	if *harPtr == false {
		if *outPtr != "stdout" {
			f, err := os.OpenFile(*outPtr, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
			if err != nil {
				panic(err)
			}

			defer f.Close()
			log.SetOutput(f)
		}
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			fmt.Println(err)
		}
		log.Println(string(requestDump))
	} else {
		if *outPtr == "stdout" {
			*outPtr = "out"
		}
		filename := fmt.Sprintf("%s.har", *outPtr)
		f, err := os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
		if err != nil {
			panic(err)
		}

		defer f.Close()
		log.SetOutput(f)
		harEntry := new(HarEntry)
		harEntry.Request = parseRequest(r)
		harEntry.StartedDateTime = time.Now()
		harEntry.Response = nil
		harEntry.Time = time.Now().Unix()

		har.addEntry(*harEntry)

		formattedHar, err := json.MarshalIndent(har, "", "\t")
		if err != nil {
			panic(err)
		}
		log.Println(string(formattedHar))
	}

	responseCode, err := strconv.Atoi(r.URL.Path[1:])
	if err == nil && responseCode < 599 {
		w.WriteHeader(responseCode)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func main() {
	portPtr := flag.Int("port", 8080, "Port number")
	outPtr = flag.String("out", "stdout", "Output location. Defaults to stdout")
	harPtr = flag.Bool("har", false, "Transforms request into valid HAR format")

	flag.Parse()

	if *harPtr == true {
		har = newHarLog()
	}
	log.SetFlags(0)

	http.HandleFunc("/", handler)

	portAddress := fmt.Sprintf(":%d", *portPtr)
	fmt.Println("Server started at port", portAddress)
	fmt.Println(http.ListenAndServe(portAddress, nil))
}
