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
	"encoding/json"
	"net/http"
	"time"
)

// Error code contained in an ansewer to the client
type AnswerErrorCode uint32

// All known error codes for an answer to the client
const (
	ANSWER_ERROR_CODE_AUTH_FAILURE                       = http.StatusUnauthorized
	ANSWER_ERROR_CODE_NOT_ENOUGHT_SPACE                  = http.StatusRequestEntityTooLarge
	ANSWER_ERROR_CODE_TERMS_NOT_ACCEPTED                 = http.StatusExpectationFailed
	ANSWER_ERROR_CODE_INTERNAL_ERROR                     = http.StatusInternalServerError
)

// The error send back to the client
type AnswerError struct {
	// The error code
	Code AnswerErrorCode `json:"code"`

	// The error message
	Message string `json:"message"`
}

// The answer send back to the client
type Answer struct {

	// The error
	Error *AnswerError `json:"error"`

	// The ID of the upload
	UploadId string `json:"upload_id"`

	// The user of the upload
	UploadUser string `json:"upload_user"`

	// The filename of the upload
	UploadFilename string `json:"upload_file"`

	// The size of the upload
	UploadSize uint64 `json:"upload_size"`

	// The time when the upload expires
	UploadExpiration time.Time `json:"upload_expiration"`
}

// Write the anser to the given response writer
func (a *Answer) Send(response http.ResponseWriter) (err error) {
	// Write response header with error code from anser
	response.Header().Set("Content-Type", "application/json; charset=utf-8")

	// Write header
	if a.Error != nil {
		// Use error code as status code OK if answer is not a success
		response.WriteHeader(int(a.Error.Code))

	} else {
		// Use status code OK if answer is a success
		response.WriteHeader(http.StatusOK)
	}

	// Create JSON encode and write anser to client
	var encoder *json.Encoder = json.NewEncoder(response)
	err = encoder.Encode(a)

	return
}
