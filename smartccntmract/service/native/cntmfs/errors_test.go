/*
 * Copyright (C) 2018 The cntmology Authors
 * This file is part of The cntmology library.
 *
 * The cntmology is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntmology is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * alcntm with The cntmology.  If not, see <http://www.gnu.org/licenses/>.
 */

package cntmfs

import (
	"fmt"
	"testing"
)

func TestErrors_Serialization(t *testing.T) {
	var e Errors
	e.AddObjectError("file1", "transfer error1")
	e.AddObjectError("file2", "transfer error2")
	e.AddObjectError("file3", "transfer error3")

	data := e.ToString()
	fmt.Printf("%v\n", data)

	var f Errors
	f.FromString(data)
	for obj, err := range f.ObjectErrors {
		fmt.Printf("obj:%s   error: %s\n", obj, err)
	}
}
