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
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
)

type UploadRequest struct {
	Title       string
	Description string

	Username string
	Password string

	Filename string

	TermsAccepted bool

	Temporary *box.Temporary
}

func ParseUploadRequest(request *http.Request) (upload *UploadRequest, err error) {
	var reader *multipart.Reader

	// Get the reader for the parts of the HTTP request
	reader, err = request.MultipartReader()
	if err != nil {
		return
	}

	// Build the upload request
	upload = &UploadRequest{}

	// Parse the multipart parts
	for {
		var part *multipart.Part

		// Get the next part
		part, err = reader.NextPart()
		if err == io.EOF {
			err = nil
			break

		} else if err != nil {
			return
		}

		// Delegate the parsing depending on the form field name
		switch part.FormName() {
		case "title":
			err = upload.parseUploadRequestTitle(part)

		case "description":
			err = upload.parseUploadRequestDescription(part)

		case "username":
			err = upload.parseUploadRequestUsername(part)

		case "password":
			err = upload.parseUploadRequestPassword(part)

		case "terms_accepted":
			err = upload.parseUploadRequestTermsAccepted(part)

		case "file":
			err = upload.parseUploadRequestFile(part)
		}

		if err != nil {
			return
		}
	}

	return
}

func parseString(part *multipart.Part) (s string, err error) {
	var data []byte

	data, err = ioutil.ReadAll(part)
	if err != nil {
		return
	}

	s = string(data)

	return
}

func (upload *UploadRequest) parseUploadRequestTitle(part *multipart.Part) (err error) {
	upload.Title, err = parseString(part)

	return
}

func (upload *UploadRequest) parseUploadRequestDescription(part *multipart.Part) (err error) {
	upload.Description, err = parseString(part)

	return
}

func (upload *UploadRequest) parseUploadRequestUsername(part *multipart.Part) (err error) {
	upload.Username, err = parseString(part)

	return
}

func (upload *UploadRequest) parseUploadRequestPassword(part *multipart.Part) (err error) {
	upload.Password, err = parseString(part)

	return
}

func (upload *UploadRequest) parseUploadRequestTermsAccepted(part *multipart.Part) (err error) {
	upload.TermsAccepted = false

	return
}

func (upload *UploadRequest) parseUploadRequestFile(part *multipart.Part) (err error) {
	// Store the filename of the upload
	upload.Filename = part.FileName()

	// Create temporary storage space for uploaded file
	upload.Temporary, err = storage.NewTemporary()
	if err != nil {
		return
	}

	// Copy the data from the part to the temporary file
	_, err = io.Copy(upload.Temporary, part)
	if err != nil {
		return
	}

	return
}

func (upload *UploadRequest) Close() (err error) {
	if upload.Temporary != nil {
		err = upload.Temporary.Close()
	}

	return
}
