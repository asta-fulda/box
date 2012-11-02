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
 
package box

import (
	"crypto/sha1"
	"flag"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"os"
	"path"
)

// Comand line arguments
var (
	flag_st_data string
	flag_st_temp string
)

func init() {
	flag.StringVar(&flag_st_data, "st_data", "/var/lib/box", "storage data")
	flag.StringVar(&flag_st_temp, "st_temp", "/var/tmp/box", "storage temp")
}

// Wrapper for the defined storage
type Storage struct {
	data string
	temp string
}

// Wrapper for a temporary file representing a currently in progress upload
type Temporary struct {
	storage *Storage

	// The temporary backend file
	file *os.File

	// Hash builder used to calculate the has while writing the file
	hasher hash.Hash

	// Writer used to write the file and calculate the hash at once
	writer io.Writer
	
	// The size of the data written to the temporary
	size uint64
}

// Connects to a storage
func ConnectStorage() (st *Storage, err error) {

	// Check if data exists and is directory
	var dataStat os.FileInfo

	dataStat, err = os.Stat(flag_st_data)
	if err != nil {
		return
	}

	if !dataStat.IsDir() {
		err = fmt.Errorf("'%s' is not a directory")
		return
	}

	// Check if temp exists and is directory
	var tempStat os.FileInfo

	tempStat, err = os.Stat(flag_st_temp)
	if err != nil {
		return
	}

	if !tempStat.IsDir() {
		err = fmt.Errorf("'%s' is not a directory")
		return
	}

	// Create storage wrapper
	st = &Storage{
		data: flag_st_data,
		temp: flag_st_temp,
	}

	return
}

// Returns the data of the storage
func (st *Storage) Data() string {
	return st.data
}

// Returns the temp of the storage
func (st *Storage) Temp() string {
	return st.temp
}

// Create a new temporary storage space
func (st *Storage) NewTemporary() (tmp *Temporary, err error) {
	tmp = &Temporary{
		storage: st,
	}

	// Create the temporary file
	tmp.file, err = ioutil.TempFile(st.temp, "upload_")
	if err != nil {
		return
	}

	// Create the hash builder
	tmp.hasher = sha1.New()

	// Create the writer
	tmp.writer = io.MultiWriter(tmp.file, tmp.hasher)

	return
}

func (st *Storage) Remove(id string) (err error) {
	// Build the full path name
	var path string = path.Join(st.data, id)

	// Remove the file
	err = os.Remove(path)

	return
}

// Close the storage
func (st *Storage) Close() (err error) {
	return
}

// Implement writer interface for temporary
func (tmp *Temporary) Write(p []byte) (n int, err error) {
  // Write the data to the combined writer
	n, err = tmp.writer.Write(p)
	
	// Update the number of written bytes
	tmp.size += uint64(n)
	
	return
}

// Get the size of the temporary file
func (tmp *Temporary) GetSize() (uint64) {
  return tmp.size
}

// Settle the temporary file and make it permanent
// This wil close the temporary file and move it to the permanent storage space
// allowing now further modification of the file
func (tmp *Temporary) Settle() (id string, err error) {
	// Close the file
	err = tmp.file.Close()
	if err != nil {
		return
	}

	// Calculate the id of the file
	id = fmt.Sprintf("%x", tmp.hasher.Sum(nil))

	// Move the temporary file to the persistent storage
	err = os.Rename(tmp.file.Name(), path.Join(tmp.storage.data, id))
	if err != nil {
		return
	}

	return
}

// Close the temporary and remove it
func (tmp *Temporary) Close() (err error) {
	defer os.Remove(tmp.file.Name())

	err = tmp.file.Close()

	return
}
