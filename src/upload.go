/*                                                                                                                                                                                                                  
 * Copyright 2011 Dustin Frisch<fooker@lab.sh>
 * 
 * This file is part of box.
 * 
 * box is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 * 
 * box is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 * 
 * You should have received a copy of the GNU General Public License
 * along with box. If not, see <http://www.gnu.org/licenses/>.
 */

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
    "encoding/json"
)

// Comand line arguments
var flag_laddr string

var flag_storage string
var flag_tempdir string

type Result struct {
    Id string
}

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
	
	var result Result
	
	fmt.Errorf("Doing...")

	// Parse form data
	log.Println("Start parsing...")
	err = request.ParseMultipartForm(4096)
	if err != nil {
		http.Error(response, err.Error(), 500)
		return
	}
	log.Println("Finished parsing...")

	// Create temporary file for upload
	tempFile, err := ioutil.TempFile(flag_tempdir, "upload_")
	if tempFile == nil || err != nil {
		http.Error(response, err.Error(), 500)
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
		http.Error(response, err.Error(), 500)
		return
	}

	// Copy content from upload to writer
	_, err = io.Copy(writer, uploadFile)
	if err != nil {
		http.Error(response, err.Error(), 500)
		return
	}

	// Get hex formated hash
	result.Id = fmt.Sprintf("%x", hasher.Sum(nil))

	// Move temporary to final storage
	err = os.Rename(tempFile.Name(), fmt.Sprintf("%s/%s", flag_storage, result.Id))
	if err != nil {
		http.Error(response, err.Error(), 500)
		return
	}
    
    // Write response
	response.Header().Add("Content-Type", "application/json")
	response.WriteHeader(200)
	
	encoder := json.NewEncoder(response)
	encoder.Encode(result)
}

func main() {
	var err error
	var listener net.Listener

	log.Println("Starting the box...")
	
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
