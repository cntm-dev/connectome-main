/*
 * Copyright (C) 2018 The cntm Authors
 * This file is part of The cntm library.
 *
 * The cntm is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Lesser General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * The cntm is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU Lesser General Public License for more details.
 *
 * You should have received a copy of the GNU Lesser General Public License
 * along with The cntm.  If not, see <http://www.gnu.org/licenses/>.
 */

package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatCntg(t *testing.T) {
	assert.Equal(t, "1", FormatCntg(1000000000))
	assert.Equal(t, "1.1", FormatCntg(1100000000))
	assert.Equal(t, "1.123456789", FormatCntg(1123456789))
	assert.Equal(t, "1000000000.123456789", FormatCntg(1000000000123456789))
	assert.Equal(t, "1000000000.000001", FormatCntg(1000000000000001000))
	assert.Equal(t, "1000000000.000000001", FormatCntg(1000000000000000001))
}

func TestParseCntg(t *testing.T) {
	assert.Equal(t, uint64(1000000000), ParseCntg("1"))
	assert.Equal(t, uint64(1000000000000000000), ParseCntg("1000000000"))
	assert.Equal(t, uint64(1000000000123456789), ParseCntg("1000000000.123456789"))
	assert.Equal(t, uint64(1000000000000000100), ParseCntg("1000000000.0000001"))
	assert.Equal(t, uint64(1000000000000000001), ParseCntg("1000000000.000000001"))
	assert.Equal(t, uint64(1000000000000000001), ParseCntg("1000000000.000000001123"))
}

func TestFormatCntm(t *testing.T) {
	assert.Equal(t, "0", FormatCntm(0))
	assert.Equal(t, "1", FormatCntm(1))
	assert.Equal(t, "100", FormatCntm(100))
	assert.Equal(t, "1000000000", FormatCntm(1000000000))
}

func TestParseCntm(t *testing.T) {
	assert.Equal(t, uint64(0), ParseCntm("0"))
	assert.Equal(t, uint64(1), ParseCntm("1"))
	assert.Equal(t, uint64(1000), ParseCntm("1000"))
	assert.Equal(t, uint64(1000000000), ParseCntm("1000000000"))
	assert.Equal(t, uint64(1000000), ParseCntm("1000000.123"))
}

func TestGenExportBlocksFileName(t *testing.T) {
	name := "blocks.dat"
	start := uint32(0)
	end := uint32(100)
	fileName := GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks_0_100.dat", fileName)
	name = "blocks"
	fileName = GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks_0_100", fileName)
	name = "blocks."
	fileName = GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks_0_100.", fileName)
	name = "blocks.export.dat"
	fileName = GenExportBlocksFileName(name, start, end)
	assert.Equal(t, "blocks.export_0_100.dat", fileName)
}
