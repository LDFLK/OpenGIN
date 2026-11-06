// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package graphschema

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// Parse decodes a graph schema document from JSON.
//
// Decoding is strict: unknown fields are rejected (so typos and unsupported keys such as
// "category" or loader fields like "file" are reported rather than silently dropped), and
// anything after the top-level JSON object is an error. The document must be a JSON object;
// a top-level null is rejected. Parse only checks that the document is well-formed; use
// Validate to check the schema rules.
func Parse(src []byte) (*Schema, error) {
	decoder := json.NewDecoder(bytes.NewReader(src))
	decoder.DisallowUnknownFields()

	// Decoding into a pointer leaves it nil for a JSON null, so null can be told apart from an
	// object.
	var schema *Schema
	if err := decoder.Decode(&schema); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, fmt.Errorf("invalid graph schema: document is empty")
		}
		return nil, fmt.Errorf("invalid graph schema: %w", err)
	}
	if schema == nil {
		return nil, fmt.Errorf("invalid graph schema: document must be a JSON object, not null")
	}

	// Only whitespace may follow the schema object.
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return nil, fmt.Errorf("invalid graph schema: unexpected data after the JSON object")
	}

	return schema, nil
}
