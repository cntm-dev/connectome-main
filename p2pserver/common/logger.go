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

package common

import "github.com/cntmio/cntmology/common/log"

type Logger interface {
	Debug(a ...interface{})
	Info(a ...interface{})
	Warn(a ...interface{})
	Error(a ...interface{})
	Fatal(a ...interface{})
	Debugf(format string, a ...interface{})
	Infof(format string, a ...interface{})
	Warnf(format string, a ...interface{})
	Errorf(format string, a ...interface{})
	Fatalf(format string, a ...interface{})
}

// the global log.Log singletion will reinit periodically
// so must be accessed by function like log.Warn
type GlobalLoggerWrapper struct{}

func NewGlobalLoggerWrapper() Logger {
	return &GlobalLoggerWrapper{}
}

func (self *GlobalLoggerWrapper) Debug(a ...interface{}) {
	log.Debug(a...)
}

func (self *GlobalLoggerWrapper) Info(a ...interface{}) {
	log.Info(a...)
}

func (self *GlobalLoggerWrapper) Warn(a ...interface{}) {
	log.Warn(a...)
}

func (self *GlobalLoggerWrapper) Error(a ...interface{}) {
	log.Error(a...)
}

func (self *GlobalLoggerWrapper) Fatal(a ...interface{}) {
	log.Fatal(a...)
}

func (self *GlobalLoggerWrapper) Debugf(format string, a ...interface{}) {
	log.Debugf(format, a...)
}

func (self *GlobalLoggerWrapper) Infof(format string, a ...interface{}) {
	log.Infof(format, a...)
}

func (self *GlobalLoggerWrapper) Warnf(format string, a ...interface{}) {
	log.Warnf(format, a...)
}

func (self *GlobalLoggerWrapper) Errorf(format string, a ...interface{}) {
	log.Errorf(format, a...)
}

func (self *GlobalLoggerWrapper) Fatalf(format string, a ...interface{}) {
	log.Fatalf(format, a...)
}

type withCcntmext struct {
	ccntmext string
	logger  Logger
}

func LoggerWithCcntmext(logger Logger, ccntmext string) Logger {
	return &withCcntmext{ccntmext: ccntmext, logger: logger}
}

func (self *withCcntmext) Debug(a ...interface{}) {
	if self.ccntmext != "" {
		t := []interface{}{self.ccntmext}
		a = append(t, a...)
	}
	self.logger.Debug(a...)
}
func (self *withCcntmext) Info(a ...interface{}) {
	if self.ccntmext != "" {
		t := []interface{}{self.ccntmext}
		a = append(t, a...)
	}
	self.logger.Info(a...)
}
func (self *withCcntmext) Warn(a ...interface{}) {
	if self.ccntmext != "" {
		t := []interface{}{self.ccntmext}
		a = append(t, a...)
	}
	self.logger.Warn(a...)
}
func (self *withCcntmext) Error(a ...interface{}) {
	if self.ccntmext != "" {
		t := []interface{}{self.ccntmext}
		a = append(t, a...)
	}
	self.logger.Error(a...)
}
func (self *withCcntmext) Fatal(a ...interface{}) {
	if self.ccntmext != "" {
		t := []interface{}{self.ccntmext}
		a = append(t, a...)
	}
	self.logger.Fatal(a...)
}

func (self *withCcntmext) Debugf(format string, a ...interface{}) {
	if self.ccntmext != "" {
		format = "%s" + format
		t := []interface{}{self.ccntmext}
		a = append(t, a...)
	}
	self.logger.Debugf(format, a...)
}

func (self *withCcntmext) Infof(format string, a ...interface{}) {
	if self.ccntmext != "" {
		format = "%s" + format
		t := []interface{}{self.ccntmext}
		a = append(t, a...)
	}
	self.logger.Infof(format, a...)
}

func (self *withCcntmext) Warnf(format string, a ...interface{}) {
	if self.ccntmext != "" {
		format = "%s" + format
		t := []interface{}{self.ccntmext}
		a = append(t, a...)
	}
	self.logger.Warnf(format, a...)
}

func (self *withCcntmext) Errorf(format string, a ...interface{}) {
	if self.ccntmext != "" {
		format = "%s" + format
		t := []interface{}{self.ccntmext}
		a = append(t, a...)
	}
	self.logger.Errorf(format, a...)
}

func (self *withCcntmext) Fatalf(format string, a ...interface{}) {
	if self.ccntmext != "" {
		format = "%s" + format
		t := []interface{}{self.ccntmext}
		a = append(t, a...)
	}
	self.logger.Fatalf(format, a...)
}
