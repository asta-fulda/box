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
	"flag"
	"fmt"
	"io"
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

	// Create storage wrapper
	st = &Storage{
		data: flag_st_data,
	}

	return
}

// Returns the data of the storage
func (st *Storage) Data() string {
	return st.data
}

func (st *Storage) copy(dst string, src string) (err error) {
	var (
		f_src *os.File
		f_dst *os.File
	)

	// Open the source
	f_src, err = os.OpenFile(src, os.O_RDONLY, 0644)
	if err != nil {
		return
	}

	defer f_src.Close()

	// Open the destination
	f_dst, err = os.OpenFile(dst, os.O_WRONLY, 0644)
	if err != nil {
		return
	}

	defer f_dst.Close()

	// Copy the content
	_, err = io.Copy(f_dst, f_src)
	if err != nil {
		return
	}
	
	return
}

func (st *Storage) move(dst string, src string) (err error) {
	err = os.Rename(src, dst)

	return
}

// Adds the file with the given filename into the storage
// If the move flag is enabled, the file will be moved to the storage by adding
// and deleting the file 
func (st *Storage) Add(id string, filename string, move bool) (err error) {
	// Build the target filename
	var target string = path.Join(st.data, id)
  
  // Copy or move the target
	if move {
		err = st.move(target, filename)
	} else {
		err = st.copy(target, filename)
	}

	return
}

// Removes the file with the given ID from the storage
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
