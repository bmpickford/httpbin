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

type Server struct {
	outfile string
	usehar  bool
	har     *HarLog
	logger  *log.Logger
}

func (server *Server) handler(w http.ResponseWriter, r *http.Request) {
	if server.usehar {
		harEntry := new(HarEntry)
		harEntry.Request = parseRequest(r)
		harEntry.StartedDateTime = time.Now()
		harEntry.Response = nil
		harEntry.Time = time.Now().Unix()

		server.har.addEntry(*harEntry)

		har := &Har{
			HarLog: *server.har,
		}
		formattedHar, err := json.MarshalIndent(har, "", "\t")
		if err != nil {
			panic(err)
		}
		server.logger.Println(string(formattedHar))
	} else {
		requestDump, err := httputil.DumpRequest(r, true)
		if err != nil {
			fmt.Println(err)
		}
		server.logger.Println(string(requestDump))
	}

	responseCode, err := strconv.Atoi(r.URL.Path[1:])
	if err == nil && responseCode < 599 {
		w.WriteHeader(responseCode)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (server *Server) logMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var f *os.File
		var err error

		if server.usehar {
			var filename = server.outfile
			if server.outfile != os.Stdout.Name() {
				filename = fmt.Sprintf("%s.har", server.outfile)
			}
			f, err = os.OpenFile(filename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0664)
			if err != nil {
				panic(err)
			}
		} else if server.outfile != os.Stdout.Name() {
			f, err = os.OpenFile(server.outfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0664)
			if err != nil {
				panic(err)
			}
		}
		defer f.Close()
		server.logger.SetFlags(0)
		server.logger.SetOutput(f)
		next.ServeHTTP(w, r)
	})
}

func main() {
	portPtr := flag.Int("port", 8080, "Port number")
	outPtr := flag.String("out", os.Stdout.Name(), "Output location. Defaults to stdout")
	harPtr := flag.Bool("har", false, "Transforms request into valid HAR format")
	flag.Parse()

	server := &Server{
		outfile: *outPtr,
		usehar:  *harPtr,
		har:     newHarLog(),
		logger:  log.Default(),
	}

	http.HandleFunc("/", server.logMiddleware(server.handler))
	portAddress := fmt.Sprintf(":%d", *portPtr)
	fmt.Println("Server started at port", portAddress)
	fmt.Println(http.ListenAndServe(portAddress, nil))
}
