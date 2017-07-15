package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gilcrest/go-API-template/pkg/config/db"
	"github.com/gilcrest/go-API-template/pkg/config/env"

	"net/http/httputil"

	"github.com/gilcrest/go-API-template/pkg/handlers"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

func main() {

	// Initializes "environment" object to be passed around to functions
	// func envInit() *env.Env
	env := envInit()
	uh := &handlers.UserHandler{Env: env}

	// create a new mux (multiplex) router
	// func NewRouter() *Router
	r := mux.NewRouter()

	// API may have multiple versions and the matching may get a bit
	// lengthy, this routeMatcher function helps with organizing that
	// func routeMatcher(rtr *mux.Router) *mux.Router
	r = handlers.PathMatcher(uh, r)

	// handle all requests with the Gorilla router by adding
	// r to the DefaultServeMux
	// func Handle(pattern string, handler Handler)
	http.Handle("/", r)

	// func ListenAndServe(addr string, handler Handler) error
	if err := http.ListenAndServe("127.0.0.1:8080", nil); err != nil {
		log.Fatal(err)
	}
}

// Initializes "environment" object to be passed around to functions
func envInit() *env.Env {

	logger, _ := zap.NewProduction()

	// returns an open database handle of 0 or more underlying connections
	// func NewDB() (*sql.DB, error)
	sqldb, err := db.NewDB()

	if err != nil {
		log.Fatal(err)
	}

	environment := &env.Env{Db: sqldb, Logger: logger}

	return environment

}

func prettyPrintRequest(req *http.Request) {
	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))
}

//func logRequest(req *http.Request) {
//
//	// type Request struct {
//	//            Method string
//	//            URL *url.URL
//	//            Proto      string // "HTTP/1.0"
//	//            ProtoMajor int    // 1
//	//            ProtoMinor int    // 0
//	//            Header Header
//	//            Body io.ReadCloser
//	//            ContentLength int64
//	//            TransferEncoding []string
//	//            Close bool
//	//            Host string
//	//            Form url.Values
//	//            PostForm url.Values
//	//            MultipartForm *multipart.Form
//	//            Trailer Header
//	//            RemoteAddr string
//	//            RequestURI string
//	//            TLS *tls.ConnectionState
//	//    }
//
//	logger := environment.Logger
//	logger.Info("Request received",
//		zap.String("URL Path", req.URL.Path[1:]),
//		zap.String("HTTP method", req.Method),
//		zap.String("URL", req.URL.String()),
//		zap.String("Protocol", req.Proto),
//		zap.Int("ProtoMajor", req.ProtoMajor),
//		zap.Int("ProtoMinor", req.ProtoMinor),
//
//		//TODO - finish logging the rest of the request
//		//fmt.Fprintf(w, "Header = %s\n", req.Header)
//		//fmt.Fprintf(w, "Body = %s\n", req.Body)
//		//fmt.Fprintf(w, "Content Length = %d\n", req.ContentLength)
//		//fmt.Fprintf(w, "Transfer Encoding = %s\n", req.TransferEncoding)
//		//fmt.Fprintf(w, "Close boolean = %t\n", req.Close)
//		//fmt.Fprintf(w, "Host = %s\n", req.Host)
//		//fmt.Fprintf(w, "Post Form Values = %s\n", req.Form)
//	)
//}
