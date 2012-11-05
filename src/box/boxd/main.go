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
	"net"
	"net/http"
	"net/http/fcgi"
	"os"
	"strconv"
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
	flag.StringVar(&flag_srv_addr, "srv_addr", "127.0.0.1:9000", "interface:port tuple to listen on (default uses fgci provided socket)")
	flag.DurationVar(&flag_quota_time, "quota_time", 14*24*time.Hour, "the time how long an upload stays alive before it expires")
	flag.Uint64Var(&flag_quota_space, "quota_space", 20*1024*1024*1024, "the maximum disk space a single user can fill up")
}

func handleRequest(response http.ResponseWriter, request *http.Request) {
	var err error

	var (
		// The transaction used for the upload
		transaction *box.Transaction

		// The record to creatre for the upload
		record box.UploadRecord

		// The answer send back to the client
		answer Answer

		// The space consumed by the user after the upload
		space_consumption uint64
	)

	// Parse the data from the POST request
	box.LogDebug("Request: %+v", request)

	err = request.ParseMultipartForm(10 * 1024)
	if err != nil {
		goto InternalError
	}

	// Open a new transaction
	transaction, err = database.BeginTransaction()
	if err != nil {
		goto InternalError
	}

	defer transaction.Rollback()

	// Fill in upload record from the posted values
	record.Filename = request.FormValue("file.name")

	record.Title = request.FormValue("title")
	record.Description = request.FormValue("description")

	// Check if terms are accepted
	if request.FormValue("terms_accepted") != "on" {
		answer.Error = &AnswerError{
      Code:    ANSWER_ERROR_CODE_TERMS_NOT_ACCEPTED,
      Message: "Terms not accepted",
    }

		goto End
	}

	// Fill in the username
	record.User = request.FormValue("username")

	// Fill in creation and expiration time
	record.Creation = time.Now()
	record.Expiration = record.Creation.Add(flag_quota_time)

	// Fill in size of the upload
	record.Size, err = strconv.ParseUint(request.FormValue("file.size"), 10, 64)
	if err != nil {
		goto InternalError
	}

	// Fill in ID of the upload
	record.Id = request.FormValue("file.hash")

	// Insert upload record into database
	box.LogDebug("Record: %+v", record)

	err = transaction.Insert(record)
	if err != nil {
		goto InternalError
	}

	// Fetch currently consumed space for user
	space_consumption, err = transaction.QuerySpaceConsumptionFor(record.User)
	if err != nil {
		goto InternalError
	}

	// Check if user has enough free space - for the real file size
	if space_consumption > flag_quota_space {
		answer.Error = &AnswerError{
			Code:    ANSWER_ERROR_CODE_NOT_ENOUGHT_SPACE,
			Message: "Not enought space",
		}

		goto End
	}

	// Move the file from the upload to the storage
	err = storage.Add(record.Id, request.FormValue("file.path"), true)
	if err != nil {
		goto InternalError
	}

	// Commit the upload - we are done
	err = transaction.Commit()
	if err != nil {
		goto InternalError
	}

	// Build final answer for client
	answer.UploadId = record.Id
	answer.UploadUser = record.User
	answer.UploadFilename = record.Filename
	answer.UploadSize = record.Size
	answer.UploadExpiration = record.Expiration

	goto End

InternalError:
	// Return a internal error
	box.LogError("Internal error: %v", err)

	answer.Error = &AnswerError{
		Code:    ANSWER_ERROR_CODE_INTERNAL_ERROR,
		Message: err.Error(),
	}

End:
	// Send the anser to the client
	box.LogDebug("Answer: %+v", answer)

	answer.Send(response)

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
		box.LogFatal("%v", err)
	}

	defer database.Close()

	// Connect to storage
	storage, err = box.ConnectStorage()
	if err != nil {
		box.LogFatal("%v", err)
	}

	defer storage.Close()

	// Dump runtime information
	box.LogInfo("Starting boxd")
	box.LogDebug("  Storage  - Data: %s", storage.Data())
	box.LogDebug("  Database - Host: %s", database.Host())
	box.LogDebug("  Database - Port: %s", database.Port())
	box.LogDebug("  Database - Name: %s", database.Name())
	box.LogDebug("  Database - User: %s", database.User())

	// Create web server communication channel
	if flag_srv_addr == "" {
		var socket *os.File

		// Create listener to FCGI socket
		socket = os.NewFile(0, "")
		listener, err = net.FileListener(socket)
		if err != nil {
			box.LogFatal("%v", err)
		}

		defer socket.Close()
		defer listener.Close()

	} else {
		// Create listener to network socket
		listener, err = net.Listen("tcp", flag_srv_addr)
		if err != nil {
			box.LogFatal("%v", err)
		}

		defer listener.Close()
	}

	// Build handler for upload requests
	handler := http.HandlerFunc(handleRequest)

	// Start serving
	err = fcgi.Serve(listener, handler)
	if err != nil {
		box.LogFatal("%v", err)
	}
}
