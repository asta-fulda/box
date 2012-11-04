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
)

func init() {
	RegisterCommand("space_consumption", space_consumption)
}

func space_consumption(database *box.Database, arguments []string) (err error) {

	// Open database transaction
	var transaction *box.Transaction

	transaction, err = database.BeginTransaction()
	if err != nil {
		return
	}

	defer transaction.Rollback()

	// Handle arguments
	switch len(arguments) {
	case 0:
		var space map[string]uint64

		// Execute the query
		space, err = transaction.QuerySpaceConsumption()
		if err != nil {
      return
    }

		// Print result
		for user, space := range space {
			fmt.Printf("%s %d\n",
				user,
				space)
		}

	case 1:
		var user string = arguments[0]
		var space uint64

		// Execute the query
		space, err = transaction.QuerySpaceConsumptionFor(user)
		if err != nil {
      return
    }

		// Print result
		fmt.Printf("%s %d\n",
			user,
			space)

	default:
		err = fmt.Errorf("To many arguments. Usage: box_util space_consumption [USER]")
	}

	return
}
