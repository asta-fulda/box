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
	"fmt"
//	"math/rand"
//	"time"
)

func main() {
	var err error
	
	box.ParseFlags()

	// Database connection
	var database *box.Database

	// Connect to database
	database, err = box.ConnectDatabase()
	if err != nil {
    box.LogFatal("%v", err)
	}

	defer database.Close()
	
	// Open transaction
	var transaction *box.Transaction
	
	transaction, err = database.BeginTransaction()
	if err != nil {
    box.LogFatal("%v", err)
  }
  
  defer transaction.Rollback()

	//	for {
	//		// Generate a random string
	//		var raw_id []byte = make([]byte, 16)
	//		for i := 0; i < 16; i++ {
	//			raw_id[i] = byte(rand.Int())
	//		}
	//
	//		id := fmt.Sprintf("%x", raw_id)
	//
	//		t := time.Unix(int64(rand.Int31()), rand.Int63())
	//
	//		// Insert some elements
	//		record := box.UploadRecord{
	//			Id: id,
	//
	//			Filename:    fmt.Sprintf("%s.txt", id),
	//			Title:       fmt.Sprintf("title_%s", id),
	//			Description: fmt.Sprintf("description_%s", id),
	//
	//			User: fmt.Sprintf("fd%04d", rand.Intn(9999)),
	//
	//			Creation:   t,
	//			Expiration: t.Add(14 * 24 * time.Hour),
	//
	//			Size: rand.Int63(),
	//		}
	//
	//		err = database.Insert(record)
	//		if err != nil {
	//			log.Fatal(err)
	//		}
	//
	//		fmt.Printf("%+v\n", record)
	//	}

	size, err := transaction.QuerySpaceConsumptionFor("asd")
	if err != nil {
    box.LogFatal("%v", err)
	}

	fmt.Printf("%+v", size)
}
