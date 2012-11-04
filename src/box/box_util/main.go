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
)

// Callback for commands
type Command func(database *box.Database, arguments []string) (err error)

// Map between command name and command - used to find the command to call for a
// given command line parameter
var commands map[string]Command = make(map[string]Command)

// Registers a tool as a command line tool
func RegisterCommand(name string, command Command) {
	commands[name] = command
}

func main() {
	var err error
  
  box.ParseFlags()

	if flag.NArg() < 1 {
		box.LogFatal("Not enought arguments. Specify utility command to execute.")
	}

	// Find the command to call
	var command_name string = flag.Arg(0)
	var command Command = commands[command_name]
	if command == nil {
		box.LogFatal("Command '%s' not found.", command_name)
	}

	// Connect to database
	var database *box.Database
	database, err = box.ConnectDatabase()
	if err != nil {
		box.LogFatal("%v", err)
	}

	defer database.Close()

	// Call the command
	err = command(database, flag.Args()[1:])
	if err != nil {
    box.LogFatal("%v", err)
  }
}
