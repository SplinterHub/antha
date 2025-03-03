// /logger/factory.go: Part of the Antha language
// Copyright (C) 2015 The Antha authors. All rights reserved.
//
// This program is free software; you can redistribute it and/or
// modify it under the terms of the GNU General Public License
// as published by the Free Software Foundation; either version 2
// of the License, or (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program; if not, write to the Free Software
// Foundation, Inc., 51 Franklin Street, Fifth Floor, Boston, MA  02110-1301, USA.
//
// For more information relating to the software or licensing issues please
// contact license@antha-lang.org or write to the Antha team c/o
// Synthace Ltd. The London Bioscience Innovation Centre
// 1 Royal College St, London NW1 0NH UK

package logger

import "sync"

var _logger *Logger
var _logger_mtx = &sync.Mutex{}

//GetLogger gets a writeable synchronized version of a Singleton Logger. If a Logger has
// not been defined a NilLogger will be returned.
func GetLogger() *Logger {
	if _logger == nil {
		var logger Logger
		logger = *NewNilLogger()
		return &logger
	}
	return _logger
}

//SetLogger sets the logger instance accessible as a Singleton through the GetLogger function.
func SetLogger(l *Logger) error {
	_logger_mtx.Lock()
	_logger = l
	_logger_mtx.Unlock()
	return nil
}
