package handlers

import (
	"fmt"
	"net/http"
	"net/http/httputil"
)

// PrintRequest wraps the call to httputil.DumpRequest
func PrintRequest(req *http.Request) error {

	// func DumpRequest(req *http.Request, body bool) ([]byte, error)
	requestDump, err := httputil.DumpRequest(req, true)
	if err != nil {
		return HTTPStatusError{http.StatusBadRequest, err}
	}
	fmt.Println(string(requestDump))
	return nil
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
