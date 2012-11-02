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
    "os"
    "fmt"
)

// Application version
const APP_NAME = "box"
const APP_VERSION = "0.1"

// Comand line arguments
var (
    flag_version bool
)

func init() {
    // Add custom error handler for command line arguments
    flag.Usage = func() {
        fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS]\n", os.Args[0])
        flag.VisitAll(func(flag *flag.Flag) {
            fmt.Fprintf(os.Stderr, "\t%s\t%s (default: %s)\n", flag.Name, flag.Usage, flag.DefValue)
        })
    }

    // Define and parse command line arguments
    flag.BoolVar(&flag_version, "v", false, "print the version numer")
}

func ParseFlags() {
    flag.Parse()

    // Handle version info flag 
    if flag_version {
        fmt.Printf("%s %s", APP_NAME, APP_VERSION)
        os.Exit(1)
    }
}