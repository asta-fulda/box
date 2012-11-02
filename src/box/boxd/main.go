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
	"box"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"time"
)

// Configuration
var ttl = time.Duration(14 * 24 * time.Hour) // 14 days

// Comand line arguments
var (
	flag_srv_addr string

	flag_quota_time  time.Duration
	flag_quota_space uint64
)

// Box connections
var (
	database *box.Database

	storage *box.Storage
)

func init() {
	// Define and parse command line arguments
	//flag.StringVar(&flag_srv_addr, "srv_addr", "", "interface:port tuple to listen on (default uses fgci provided socket)")
	flag.StringVar(&flag_srv_addr, "srv_addr", "127.0.0.1:10000", "interface:port tuple to listen on (default uses fgci provided socket)")
	flag.DurationVar(&flag_quota_time, "quota_time", 14*24*time.Hour, "the time how long an upload stays alive before it expires")
	flag.Uint64Var(&flag_quota_space, "quota_space", 20*1024*1024*1024, "the maximum disk space a single user can fill up")
}

func handleRequest(response http.ResponseWriter, request *http.Request) {
	var err error

	var (
		// The upload request parsed from the HTTP request
		upload *UploadRequest

		// The transaction used for the upload
		transaction *box.Transaction

		// The record to creatre for the upload
		record box.UploadRecord

		// The answer send back to the client
		answer Answer

		// The space currently consumed by the user
		current_space_consumption uint64
	)

	// Parse the upload request
	upload, err = ParseUploadRequest(request)
	if err != nil {
		goto InternalError
	}

	defer upload.Close()

	fmt.Printf("Upload: %+v\n", upload)

	// Open a new transaction
	transaction, err = database.BeginTransaction()
	if err != nil {
		goto InternalError
	}

	defer transaction.Rollback()

	// Fill in upload record from the posted values
	record.Filename = upload.Filename
	record.Title = upload.Title
	record.Description = upload.Description

	// Check username and password
	if upload.Username != upload.Password {
		answer.Success = false
		answer.ErrorCode = AUTH_FAILURE
		answer.ErrorMessage = "Username and password do not match"
	}

	// Fill in the username
	record.User = upload.Username

	// Fill in creation and expiration time
	record.Creation = time.Now()
	record.Expiration = record.Creation.Add(flag_quota_time)

	// Fetch currently consumed space for user
	current_space_consumption, err = transaction.QuerySpaceConsumptionFor(upload.Username)
	if err != nil {
		goto InternalError
	}

	// Check if user has enough free space - for the real file size
	if current_space_consumption+upload.Temporary.GetSize() > flag_quota_space {
		answer.Success = false
		answer.ErrorCode = NOT_ENOUGHT_SPACE
		answer.ErrorMessage = "Not enought space left for upload"

		goto End
	}

	// Fill in size of the upload
	record.Size = upload.Temporary.GetSize()

	// Make the temporary persistent
	record.Id, err = upload.Temporary.Settle()
	if err != nil {
		goto InternalError
	}

	fmt.Printf("Record: %+v\n", record)

	// Insert upload record into database
	err = transaction.Insert(record)
	if err != nil {
		goto InternalError
	}

	// Commit the upload - we are done
	err = transaction.Commit()
	if err != nil {
		goto InternalError
	}
	
	// Build final answer for client
	answer.Success = true
	answer.UploadId = record.Id
	answer.UploadExpiration = record.Expiration

	goto End

InternalError:
	answer.Success = false
	answer.ErrorCode = INTERNAL_ERROR
	answer.ErrorMessage = err.Error()

End:
	// Send the anser to the client
	answer.Send(response)

	fmt.Printf("Answer: %+v\n", answer)

	return
}

func main() {
	var err error

	box.ParseFlags()

	// Listener to the socket
	var listener net.Listener

	// Connect to database
	database, err = box.ConnectDatabase()
	if err != nil {
		log.Fatal(err)
	}

	defer database.Close()

	// Connect to storage
	storage, err = box.ConnectStorage()
	if err != nil {
		log.Fatal(err)
	}

	defer storage.Close()

	// Create web server communication channel
	if flag_srv_addr == "" {
		var socket *os.File

		// Create listener to FCGI socket
		socket = os.NewFile(0, "")
		listener, err = net.FileListener(socket)
		if err != nil {
			log.Fatal(err)
		}

		defer socket.Close()
		defer listener.Close()

	} else {
		// Create listener to network socket
		listener, err = net.Listen("tcp", flag_srv_addr)
		if err != nil {
			log.Fatal(err)
		}

		defer listener.Close()
	}

	// Build handler for upload requests
	handler := http.HandlerFunc(handleRequest)

	// Start serving
	err = fcgi.Serve(listener, handler)
	if err != nil {
		log.Fatal(err)
	}
}
