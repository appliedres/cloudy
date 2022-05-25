package cloudy

import (
	"bytes"
	"fmt"
	"io"
	"sync"
)

// Simple multi error class... copied from https://github.com/olekukonko/merror/blob/master/merror.go and modified

// Error Type
type MultiErrors struct {
	lock  *sync.RWMutex
	items []error
}

// Initiate Error Instance
func MultiError() *MultiErrors {
	return &MultiErrors{new(sync.RWMutex), []error{}}
}

// Append new error
func (e *MultiErrors) Append(err ...error) {
	e.lock.Lock()
	defer e.lock.Unlock()
	for _, v := range err {
		e.items = append(e.items, v)
	}
}

// Merge Errors
func (e *MultiErrors) Merge(err *MultiErrors) {
	e.Append(err.List()...)
}

// List all error items
func (e *MultiErrors) List() []error {
	e.lock.RLock()
	defer e.lock.RUnlock()
	return e.items
}

// Clear all error items
func (e *MultiErrors) Clear() {
	e.lock.Lock()
	defer e.lock.Unlock()
	e.items = []error{}
}

// Total lent of error in list
func (e *MultiErrors) Len() int {
	e.lock.Lock()
	defer e.lock.Unlock()
	return len(e.items)
}

// Check if it has errors
func (e *MultiErrors) HasError() bool {
	return e.Len() > 0
}

// Return Stringer interface
func (e *MultiErrors) String() string {
	buf := new(bytes.Buffer)
	l := e.Len()
	for i, v := range e.List() {
		fmt.Fprintf(buf, "%s", v)
		if (i + 1) < l {
			fmt.Fprint(buf, "; ")
		}
	}
	return buf.String()
}

// Display errors in tabulated format
func (e *MultiErrors) Tab(w io.Writer) {
	fmt.Fprintf(w, "%-1s%d Error(s) Found\n", " ", e.Len())
	for _, v := range e.List() {
		fmt.Fprintf(w, "%-2s%-2s%s\n", " ", "-", v)
	}
	fmt.Fprintln(w)
}

func (e *MultiErrors) Error() string {
	if e.Len() == 0 {
		return ""
	}
	if e.Len() == 1 {
		return e.items[0].Error()
	}
	return e.String()
}

func (e *MultiErrors) AsErr() error {
	if e.Len() == 0 {
		return nil
	}
	return e
}

// Display errors in Line format
func (e *MultiErrors) Line(w io.Writer) {
	for _, v := range e.List() {
		fmt.Fprintf(w, "%s\t\n", v)
	}
	fmt.Fprintln(w)
}
