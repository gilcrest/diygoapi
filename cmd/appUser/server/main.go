package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gilcrest/go-API-template/pkg/appUser"
	"github.com/gilcrest/go-API-template/pkg/config/db"
	"github.com/gilcrest/go-API-template/pkg/config/env"

	"github.com/gorilla/mux"
)

var environment env.Env

func main() {

	envInit()

	// create a new mux router
	// func NewRouter() *Router
	r := mux.NewRouter()

	// match only POST requests on /memberlog/
	// func (r *Router) HandleFunc(path string, f func(http.ResponseWriter,
	// *http.Request)) *Route
	r.HandleFunc("/appUser/create", handleMbrLog).Methods("GET")

	// handle all requests with the Gorilla router by adding
	// r to the DefaultServeMux
	// func Handle(pattern string, handler Handler)
	http.Handle("/", r)

	//
	// func ListenAndServe(addr string, handler Handler) error
	if err := http.ListenAndServe("127.0.0.1:8080", nil); err != nil {
		log.Fatal(err)
	}
}

// sets DBCon global variable with
// an open database handle of 0 or more underlying connections
func envInit() {

	// returns an open database handle of 0 or more underlying connections
	// func NewDB() (*sql.DB, error)
	sqldb, err := db.NewDB()

	if err != nil {
		log.Fatal(err)
	}

	environment = env.Env{Db: sqldb}

}

func handleMbrLog(w http.ResponseWriter, req *http.Request) {

	// retrieve the context from the http.Request
	ctx := req.Context()
	log.Printf("handleMbrLog started")
	defer log.Printf("handleMbrLog ended")

	// db.NewContext function creates and begins a new sql.Tx, which pulls from the
	// previously opened database (postgres) connection pool and starts a database
	// transaction.  In addition, the pointer to this "started" sql.Tx is included
	// in the above created context
	ctx = db.AddDBTx2Context(ctx, environment, nil)

	inputUsr := appUser.User{Username: "repoMan", MobileID: "(617) 302-7777", Email: "repoman@alwaysintense.com", FirstName: "Otto", LastName: "Maddox"}
	//auditUsr := appUser.User{Username: "bud", MobileID: "(617) 302-7777", Email: "repoman@alwaysintense.com", FirstName: "Otto", LastName: "Maddox"}

	//
	logsWritten, err := inputUsr.Create(ctx)

	fmt.Printf("logsWritten = %d\n", logsWritten)

	tx, ok := db.DBTxFromContext(ctx)

	if ok && logsWritten > 0 {
		err = tx.Commit()
		if err != nil {
			log.Fatal(err)
		}
	} else if logsWritten <= 0 {
		log.Fatal(err)
	}

	// type Request struct {
	//            Method string
	//            URL *url.URL
	//            Proto      string // "HTTP/1.0"
	//            ProtoMajor int    // 1
	//            ProtoMinor int    // 0
	//            Header Header
	//            Body io.ReadCloser
	//            ContentLength int64
	//            TransferEncoding []string
	//            Close bool
	//            Host string
	//            Form url.Values
	//            PostForm url.Values
	//            MultipartForm *multipart.Form
	//            Trailer Header
	//            RemoteAddr string
	//            RequestURI string
	//            TLS *tls.ConnectionState
	//    }
	fmt.Fprintf(w, "Hi there, I love %s!\n", req.URL.Path[1:])
	fmt.Fprintf(w, "HTTP method = %s\n", req.Method)
	fmt.Fprintf(w, "URL = %s\n", req.URL)
	fmt.Fprintf(w, "Protocol = %s\n", req.Proto)
	fmt.Fprintf(w, "ProtoMajor = %d\n", req.ProtoMajor)
	fmt.Fprintf(w, "ProtoMinor = %d\n", req.ProtoMinor)
	fmt.Fprintf(w, "Header = %s\n", req.Header)
	fmt.Fprintf(w, "Body = %s\n", req.Body)
	fmt.Fprintf(w, "Content Length = %d\n", req.ContentLength)
	fmt.Fprintf(w, "Transfer Encoding = %s\n", req.TransferEncoding)
	fmt.Fprintf(w, "Close boolean = %t\n", req.Close)
	fmt.Fprintf(w, "Host = %s\n", req.Host)
	fmt.Fprintf(w, "Post Form Values = %s\n", req.Form)

}
