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
	"os"
)

func main() {
	var err error

	box.ParseFlags()

	// Box connections
	var (
		database *box.Database

		storage *box.Storage
	)

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

	// Open transaction
	var transaction *box.Transaction

	transaction, err = database.BeginTransaction()
	if err != nil {
		box.LogFatal("%v", err)
	}

	defer transaction.Rollback()

	// Find all expired uploads
	var ids []string

	// Get all ids in storage
	ids, err = storage.List()
	if err != nil {
		box.LogFatal("%v", err)
	}

	// Remove all expired uploads from storage
	for _, id := range ids {
		var expired bool

		// Query if the file is expired
		expired, err = transaction.QueryExpiredUpload(id)
		if err != nil {
			box.LogFatal("%v", err)
		}

		if expired {
			box.LogDebug("Removing file: %s", id)

			err = storage.Remove(id)
			if err != os.ErrNotExist {
				// File is already gone - ignore it
			} else if err != nil {
				box.LogError("%v", err)
			}
		}
	}
}
