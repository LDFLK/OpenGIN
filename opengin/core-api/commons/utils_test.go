// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package commons

import (
	"bytes"
	"testing"

	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestConvertStringToAnyIncludesProtobufFraming(t *testing.T) {
	const name = "Ministry of Health"
	packed := ConvertStringToAny(name)

	if bytes.Equal(packed.Value, []byte(name)) {
		t.Fatal("ConvertStringToAny stored raw UTF-8; expected StringValue framing")
	}
	// google.protobuf.StringValue is field 1, length-delimited: 0x0A <len> <utf8>
	if len(packed.Value) < 2 || packed.Value[0] != 0x0A || packed.Value[1] != byte(len(name)) {
		t.Fatalf("missing StringValue framing, got %x", packed.Value)
	}

	var sv wrapperspb.StringValue
	if err := packed.UnmarshalTo(&sv); err != nil {
		t.Fatalf("UnmarshalTo: %v", err)
	}
	if sv.Value != name {
		t.Fatalf("unpacked = %q, want %q", sv.Value, name)
	}

	// Contrast: labeling raw UTF-8 as StringValue (the old search bug) does not decode.
	broken := &anypb.Any{
		TypeUrl: "type.googleapis.com/google.protobuf.StringValue",
		Value:   []byte(name),
	}
	var brokenSV wrapperspb.StringValue
	if err := broken.UnmarshalTo(&brokenSV); err == nil && brokenSV.Value == name {
		t.Fatal("raw UTF-8 Any unexpectedly unmarshaled to the full name")
	}
	if ExtractStringFromAny(broken) == name {
		t.Fatal("ExtractStringFromAny should not treat unframed UTF-8 as a StringValue")
	}
}
