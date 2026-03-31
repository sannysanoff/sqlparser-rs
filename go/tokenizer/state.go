// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package tokenizer

import (
	"unicode/utf8"

	"github.com/user/sqlparser/span"
)

// State represents the tokenizer's state during parsing.
// It tracks the current position in the input string, line and column numbers.
type State struct {
	query  string
	pos    int // current byte position in query
	line   uint64
	column uint64
	end    int // length of query
}

// NewState creates a new State for tokenizing the given query
func NewState(query string) *State {
	return &State{
		query:  query,
		pos:    0,
		line:   1,
		column: 1,
		end:    len(query),
	}
}

// Next returns the next character and advances the stream.
// Returns 0 and false if at end of input.
func (s *State) Next() (rune, bool) {
	if s.pos >= s.end {
		return 0, false
	}

	r, size := utf8.DecodeRuneInString(s.query[s.pos:])
	s.pos += size

	if r == '\n' {
		s.line++
		s.column = 1
	} else {
		s.column++
	}

	return r, true
}

// Peek returns the next character without advancing the stream.
// Returns 0 and false if at end of input.
func (s *State) Peek() (rune, bool) {
	if s.pos >= s.end {
		return 0, false
	}
	r, _ := utf8.DecodeRuneInString(s.query[s.pos:])
	return r, true
}

// PeekN returns the nth character ahead without advancing.
// Returns 0 and false if beyond end of input.
func (s *State) PeekN(n int) (rune, bool) {
	pos := s.pos
	for i := 0; i < n && pos < s.end; i++ {
		r, size := utf8.DecodeRuneInString(s.query[pos:])
		if i == n-1 {
			return r, true
		}
		pos += size
	}
	return 0, false
}

// Location returns the current location (line and column)
func (s *State) Location() span.Location {
	return span.Location{
		Line:   s.line,
		Column: s.column,
	}
}

// Position returns the current byte position
func (s *State) Position() int {
	return s.pos
}

// Query returns the full query string
func (s *State) Query() string {
	return s.query
}

// TakeWhile consumes characters while the predicate returns true
// Returns the consumed string and the count of runes consumed
func (s *State) TakeWhile(predicate func(rune) bool) string {
	var result []rune
	for {
		ch, ok := s.Peek()
		if !ok || !predicate(ch) {
			break
		}
		s.Next()
		result = append(result, ch)
	}
	return string(result)
}

// TakeWhileWithNext consumes characters while the predicate returns true.
// The predicate receives both current and next character.
func (s *State) TakeWhileWithNext(predicate func(current, next rune) bool) string {
	var result []rune
	for {
		ch, ok := s.Peek()
		if !ok {
			break
		}
		nextCh, _ := s.PeekN(1)
		if !predicate(ch, nextCh) {
			break
		}
		s.Next()
		result = append(result, ch)
	}
	return string(result)
}

// SkipWhile skips characters while the predicate returns true
func (s *State) SkipWhile(predicate func(rune) bool) {
	for {
		ch, ok := s.Peek()
		if !ok || !predicate(ch) {
			break
		}
		s.Next()
	}
}

// IsEOF returns true if at end of input
func (s *State) IsEOF() bool {
	return s.pos >= s.end
}

// Remaining returns the remaining unparsed portion of the query
func (s *State) Remaining() string {
	if s.pos >= s.end {
		return ""
	}
	return s.query[s.pos:]
}

// Clone creates a copy of the current state
func (s *State) Clone() *State {
	return &State{
		query:  s.query,
		pos:    s.pos,
		line:   s.line,
		column: s.column,
		end:    s.end,
	}
}
