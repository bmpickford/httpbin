package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"os"
	"strconv"
)

type Server struct {
	outfile string
	usehar  bool
	har     *HarLog
	logger  *log.Logger
}

func (server *Server) handleHar(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if server.usehar {
			harEntry := parseEntry(r)
			server.har.addEntry(*harEntry)

			har := &Har{
				HarLog: *server.har,
			}
			server.logger.Println(har.format())
		}
		next.ServeHTTP(w, r)
	})
}

func (server *Server) handleHTTPDump(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !server.usehar {
			requestDump, err := httputil.DumpRequest(r, true)
			if err != nil {
				panic(err)
			}
			server.logger.Println(string(requestDump))
		}
		next.ServeHTTP(w, r)
	})
}

func (server *Server) setLogger(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		var f *os.File = os.Stdout
		var filename = server.outfile
		var permissions = os.O_RDWR | os.O_CREATE | os.O_APPEND

		if server.outfile != os.Stdout.Name() {
			if server.usehar {
				permissions = os.O_RDWR | os.O_CREATE | os.O_TRUNC
				filename = fmt.Sprintf("%s.har", server.outfile)
			}

			fmt.Printf("Printing to file at: %s\n", filename)

			f, err = os.OpenFile(filename, permissions, 0664)
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

func handler(w http.ResponseWriter, r *http.Request) {
	responseCode, err := strconv.Atoi(r.URL.Path[1:])
	if err == nil && responseCode < 599 {
		w.WriteHeader(responseCode)
		return
	}
	w.WriteHeader(http.StatusOK)
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

	http.HandleFunc("/", server.setLogger(server.handleHar(server.handleHTTPDump(handler))))
	portAddress := fmt.Sprintf(":%d", *portPtr)
	fmt.Println("Server started at port", portAddress)
	fmt.Println(http.ListenAndServe(portAddress, nil))
}
