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
	"os"
	"time"
)

// The log levels

type LogLevel uint

// Comand line arguments
var (
	flag_log_level uint
)

const (
	LOG_LEVEL_DEBUG LogLevel = iota
	LOG_LEVEL_INFO
	LOG_LEVEL_WARNING
	LOG_LEVEL_ERROR
	LOG_LEVEL_FATAL
)

const (
	LOG_LEVEL_NAME_DEBUG   string = "DEBUG"
	LOG_LEVEL_NAME_INFO           = "INFO"
	LOG_LEVEL_NAME_WARNING        = "WARNING"
	LOG_LEVEL_NAME_ERROR          = "ERROR"
	LOG_LEVEL_NAME_FATAL          = "FATAL"
)

func init() {
	flag.UintVar(&flag_log_level, "log_level", uint(LOG_LEVEL_DEBUG), "the minimal log level to display")
}

func logLevelName(level LogLevel) string {
	switch level {
	case LOG_LEVEL_DEBUG:
		return LOG_LEVEL_NAME_DEBUG
	case LOG_LEVEL_INFO:
		return LOG_LEVEL_NAME_INFO
	case LOG_LEVEL_WARNING:
		return LOG_LEVEL_NAME_WARNING
	case LOG_LEVEL_ERROR:
		return LOG_LEVEL_NAME_ERROR
	case LOG_LEVEL_FATAL:
		return LOG_LEVEL_NAME_FATAL
	}

	return ""
}

func Log(level LogLevel, format string, a ...interface{}) {
	// Print the log message if required
	if level >= LogLevel(flag_log_level) {
		fmt.Printf("%v | %7s | %s\n", time.Now(), logLevelName(level), fmt.Sprintf(format, a...))
	}

	// Check for fatal messages and exit
	if level >= LOG_LEVEL_FATAL {
		os.Exit(1)
	}
}

func LogDebug(format string, a ...interface{}) {
	Log(LOG_LEVEL_DEBUG, format, a)
}

func LogInfo(format string, a ...interface{}) {
	Log(LOG_LEVEL_INFO, format, a)
}

func LogWarning(format string, a ...interface{}) {
	Log(LOG_LEVEL_WARNING, format, a)
}

func LogError(format string, a ...interface{}) {
	Log(LOG_LEVEL_ERROR, format, a)
}

func LogFatal(format string, a ...interface{}) {
	Log(LOG_LEVEL_FATAL, format, a)
}
