// Copyright (c) 2013-2015 The btcsuite developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package wire

import (
	"fmt"
)

// ErrorCode describes a kind of message error.
type ErrorCode int

// These constants are used to identify a specific Error.
const (
	// ErrVarBytesTooLong is returned when a variable-length byte slice
	// exceeds the maximum message size allowed.
	ErrVarBytesTooLong ErrorCode = iota

	// ErrMsgInvalidForPVer is returned when a message is invalid for
	// the expected protocol version.
	ErrMsgInvalidForPVer

	// ErrInvalidMsg is returned for an invalid message structure.
	ErrInvalidMsg

	// ErrMalformedStrictString is returned when a string that has strict
	// formatting requirements does not conform to the requirements.
	ErrMalformedStrictString

	// ErrTooManyManyMixPairReqs is returned when the number of mix pair
	// request message hashes exceeds the maximum allowed.
	ErrTooManyManyMixPairReqs

	// ErrMixPairReqScriptClassTooLong is returned when a mixing script
	// class type string is longer than allowed by the protocol.
	ErrMixPairReqScriptClassTooLong

	// ErrTooManyMixPairReqUTXOs is returned when a MixPairReq message
	// contains more UTXOs than allowed by the protocol.
	ErrTooManyMixPairReqUTXOs

	// ErrTooManyPrevMixMsgs is returned when too many previous messages of
	// a mix run are referenced by a message.
	ErrTooManyPrevMixMsgs
)

// Map of ErrorCode values back to their constant names for pretty printing.
var errorCodeStrings = map[ErrorCode]string{
	ErrTooManyManyMixPairReqs:       "ErrTooManyManyMixPairReqs",
	ErrMixPairReqScriptClassTooLong: "ErrMixPairReqScriptClassTooLong",
	ErrTooManyMixPairReqUTXOs:       "ErrTooManyMixPairReqUTXOs",
	ErrTooManyPrevMixMsgs:           "ErrTooManyPrevMixMsgs",
}

// String returns the ErrorCode as a human-readable name.
func (e ErrorCode) String() string {
	if s := errorCodeStrings[e]; s != "" {
		return s
	}
	return fmt.Sprintf("Unknown ErrorCode (%d)", int(e))
}

// Error implements the error interface.
func (e ErrorCode) Error() string {
	return e.String()
}

// Is implements the interface to work with the standard library's errors.Is.
//
// It returns true in the following cases:
// - The target is a *MessageError and the error codes match
// - The target is an ErrorCode and it the error codes match
func (e ErrorCode) Is(target error) bool {
	// nolint: errorlint
	switch target := target.(type) {
	case *MessageError:
		return e == target.ErrorCode

	case ErrorCode:
		return e == target
	}

	return false
}

// MessageError describes an issue with a message.
// An example of some potential issues are messages from the wrong bitcoin
// network, invalid commands, mismatched checksums, and exceeding max payloads.
//
// This provides a mechanism for the caller to type assert the error to
// differentiate between general io errors such as io.EOF and issues that
// resulted from malformed messages.
type MessageError struct {
	Func        string    // Function name
	ErrorCode   ErrorCode // Describes the kind of error
	Description string    // Human readable description of the issue
}

// Error satisfies the error interface and prints human-readable errors.
func (e *MessageError) Error() string {
	if e.Func != "" {
		return fmt.Sprintf("%v: %v", e.Func, e.Description)
	}
	return e.Description
}

// messageError creates an error for the given function and description.
func messageError(f string, desc string) *MessageError {
	return &MessageError{Func: f, Description: desc}
}

// messageErrorWithCode creates an Error given a set of arguments.
func messageErrorWithCode(funcName string, c ErrorCode, desc string) *MessageError {
	return &MessageError{Func: funcName, ErrorCode: c, Description: desc}
}

// Is implements the interface to work with the standard library's errors.Is.
//
// It returns true in the following cases:
// - The target is a *MessageError and the error codes match
// - The target is an ErrorCode and it the error codes match
func (e *MessageError) Is(target error) bool {
	// nolint: errorlint
	switch target := target.(type) {
	case *MessageError:
		return e.ErrorCode == target.ErrorCode

	case ErrorCode:
		return target == e.ErrorCode
	}

	return false
}

// Unwrap returns the underlying wrapped error if it is not ErrOther.
// Unwrap returns the ErrorCode. Else, it returns nil.
func (e *MessageError) Unwrap() error {
	return e.ErrorCode
}
