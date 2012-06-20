package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"

//    "encoding/json"
)

// Comand line arguments
var flag_laddr string

var flag_storage string
var flag_tempdir string

func init() {
	// Add custom error handler for command line arguments
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
		flag.VisitAll(func(flag *flag.Flag) {
			fmt.Fprintf(os.Stderr, "\t%s\t%s (default: %s)\n", flag.Name, flag.Usage, flag.DefValue)
		})
	}

	// Define and parse command line arguments
	flag.StringVar(&flag_laddr, "laddr", "", "interface:port tuple to listen on (default uses fgci provided socket)")

	flag.StringVar(&flag_storage, "storage", "./storage", "directory of the storage")
	flag.StringVar(&flag_tempdir, "tempdir", "./tmp", "directory used to temporary store uploaded files")

	flag.Parse()
}

func upload(response http.ResponseWriter, request *http.Request) {
    var err error
    
	// Parse form data
	err = request.ParseMultipartForm(4096)
	if err != nil {
		response.WriteHeader(501)
		response.Write([]byte(err.Error()))
		return
	}

	// Create temporary file for upload
	tempFile, err := ioutil.TempFile(flag_tempdir, "upload_")
	if tempFile == nil || err != nil {
		response.WriteHeader(501)
		response.Write([]byte(err.Error()))
		return
	}

	// Close and delete file later
	defer tempFile.Close()
	defer os.Remove(tempFile.Name())

	// Create MD5Sum hash builder
	hasher := md5.New()

	// Create writer feeding the temp file and the hash builder
	writer := io.MultiWriter(tempFile, hasher)

	// Get uploaded file from form
	uploadFile, _, err := request.FormFile("file")
	if err != nil {
		response.WriteHeader(501)
		response.Write([]byte(err.Error()))
		return
	}

	// Copy content from upload to writer
	_, err = io.Copy(writer, uploadFile)
	if err != nil {
		response.WriteHeader(501)
		response.Write([]byte(err.Error()))
		return
	}

	// Get hex formated hash
	hash := fmt.Sprintf("%x", hasher.Sum(nil))

	// Move temporary to final storage
	err = os.Rename(tempFile.Name(), flag_storage+"/"+hash)
	if err != nil {
		response.WriteHeader(501)
		response.Write([]byte(err.Error()))
		return
	}

	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(200)
	response.Write([]byte(hash))
}

func main() {
	var err error
	var listener net.Listener
	
	log.Println("Settings: laddr = " + flag_laddr)
	log.Println("Settings: storage = " + flag_storage)
	log.Println("Settings: tempdir = " + flag_tempdir)

	if flag_laddr == "" {
		// Create listener to FCGI socket
		socket := os.NewFile(0, "")
		listener, err = net.FileListener(socket)
		if err != nil {
			log.Fatal(err)
		}

	} else {
		// Create listener to network socket
		listener, err = net.Listen("tcp", flag_laddr)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Build handler for upload requests
	handler := http.HandlerFunc(upload)

	// Start serving
	err = fcgi.Serve(listener, handler)
	if err != nil {
		log.Fatal(err)
	}
}
