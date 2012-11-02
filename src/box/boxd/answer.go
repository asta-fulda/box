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
	OK                AnswerErrorCode = http.StatusOK
	AUTH_FAILURE                      = http.StatusUnauthorized
	NOT_ENOUGHT_SPACE                 = http.StatusRequestEntityTooLarge
	INTERNAL_ERROR                    = http.StatusInternalServerError
)

// The answer send back to the client
type Answer struct {

	// Request was successfull
	Success bool `json:"success"`

	// Error code
	ErrorCode AnswerErrorCode `json:"error_code"`

	// Error message - nil if no error occured
	ErrorMessage string `json:"error_message"`
	
	// The ID of the upload
	UploadId string `json:"upload_id"`
	
	// The time when the upload expires
	UploadExpiration time.Time `json:"upload_expiration"`
}

// Write the anser to the given response writer
func (a *Answer) Send(response http.ResponseWriter) (err error) {
	// Write response header with error code from anser
	response.Header().Set("Content-Type", "application/json; charset=utf-8")
	response.WriteHeader(int(a.ErrorCode))

	// Create JSON encode and write anser to client
	var encoder *json.Encoder = json.NewEncoder(response)
	err = encoder.Encode(a)

	return
}
