// Copyright 2016 Michael Schenk. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package util

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
)

type ContextLogger struct {
	skip   int
	logger *log.Logger
}

func NewContextLogger(x *log.Logger, skip int) *ContextLogger {
	return &ContextLogger{logger: x, skip: skip}
}

func (c *ContextLogger) Printf(format string, v ...interface{}) {
	s := c.stack()
	v = append([]interface{}{s.String()}, v...)
	c.logger.Printf(`%s`+format, v...)
}

func (c *ContextLogger) Print(v ...interface{}) {
	s := c.stack()
	v = append([]interface{}{s.String()}, v...)
	c.logger.Print(v...)
}

func (c *ContextLogger) Println(v ...interface{}) {
	s := c.stack()
	v = append([]interface{}{s.String()}, v...)
	c.logger.Println(v...)
}

func (c *ContextLogger) stack() FamilyCallStack {
	pc := make([]uintptr, 15)
	n := runtime.Callers(c.skip, pc)
	fcs := NewFamilyCallStack(runtime.CallersFrames(pc[:n]))
	return fcs
}

func NewFamilyCallStack(f *runtime.Frames) FamilyCallStack {
	child, _ := f.Next()
	parent, _ := f.Next()

	cframe := CallFrame{
		File:     filepath.Base(child.File),
		Line:     child.Line,
		Function: filepath.Base(child.Function),
	}

	pframe := CallFrame{
		File:     filepath.Base(parent.File),
		Line:     parent.Line,
		Function: filepath.Base(parent.Function),
	}

	return FamilyCallStack{Parent: pframe, Child: cframe}
}

type CallFrame struct {
	File     string
	Line     int
	Function string
}

type FamilyCallStack struct {
	Parent CallFrame
	Child  CallFrame
}

func (f FamilyCallStack) String() string {
	return fmt.Sprint(f.Parent, " > ", f.Child, " ")
}

func (c CallFrame) String() string {
	return fmt.Sprintf("%s:%d %s", c.File, c.Line, c.Function)
}
