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

package token

// Location represents a position in the source code
type Location struct {
	Line   uint64
	Column uint64
}

// Span represents a range of source code from start to end
type Span struct {
	Start Location
	End   Location
}

// NewSpan creates a new Span from start and end locations
func NewSpan(start, end Location) Span {
	return Span{Start: start, End: end}
}

// Merge combines two spans into one that covers both
func (s Span) Merge(other Span) Span {
	start := s.Start
	if other.Start.Line < start.Line || (other.Start.Line == start.Line && other.Start.Column < start.Column) {
		start = other.Start
	}

	end := s.End
	if other.End.Line > end.Line || (other.End.Line == end.Line && other.End.Column > end.Column) {
		end = other.End
	}

	return Span{Start: start, End: end}
}

// IsValid returns true if the span has been set (non-zero)
func (s Span) IsValid() bool {
	return s.Start.Line > 0 || s.Start.Column > 0
}

// Spanned is an interface for AST nodes that track source location
type Spanned interface {
	Span() Span
}
