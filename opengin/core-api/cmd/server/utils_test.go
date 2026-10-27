// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"bytes"
	"testing"

	"lk/datafoundation/core-api/commons"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestEntityFromGraphFilterResultPacksFramedStringValue(t *testing.T) {
	const entityName = "Ministry of Health"
	entity := entityFromGraphFilterResult(map[string]interface{}{
		"id":         "minister-health-001",
		"kind":       "Organization",
		"minorKind":  "Minister",
		"created":    "2024-01-01T00:00:00Z",
		"terminated": "",
		"name":       entityName,
	})

	if entity.Name == nil || entity.Name.Value == nil {
		t.Fatal("expected name Any to be set")
	}

	wantType := "type.googleapis.com/google.protobuf.StringValue"
	if entity.Name.Value.TypeUrl != wantType {
		t.Fatalf("typeUrl = %q, want %q", entity.Name.Value.TypeUrl, wantType)
	}

	// Raw UTF-8 (the old ReadEntities bug) must not be stored in Any.Value.
	if bytes.Equal(entity.Name.Value.Value, []byte(entityName)) {
		t.Fatal("name Any.Value is raw UTF-8; expected protobuf StringValue framing")
	}

	var unpacked wrapperspb.StringValue
	if err := entity.Name.Value.UnmarshalTo(&unpacked); err != nil {
		t.Fatalf("UnmarshalTo StringValue: %v", err)
	}
	if unpacked.Value != entityName {
		t.Fatalf("unpacked name = %q, want %q", unpacked.Value, entityName)
	}

	if got := commons.ExtractStringFromAny(entity.Name.Value); got != entityName {
		t.Fatalf("ExtractStringFromAny = %q, want %q", got, entityName)
	}

	if entity.Name.StartTime != "2024-01-01T00:00:00Z" {
		t.Fatalf("StartTime = %q", entity.Name.StartTime)
	}
}

func TestEntityFromGraphFilterResultSetsTerminatedOnName(t *testing.T) {
	entity := entityFromGraphFilterResult(map[string]interface{}{
		"id":         "ended-entity",
		"kind":       "Person",
		"minorKind":  "Employee",
		"created":    "2024-01-01T00:00:00Z",
		"terminated": "2024-12-31T00:00:00Z",
		"name":       "Ended",
	})
	if entity.Terminated != "2024-12-31T00:00:00Z" {
		t.Fatalf("Terminated = %q", entity.Terminated)
	}
	if entity.Name.EndTime != "2024-12-31T00:00:00Z" {
		t.Fatalf("Name.EndTime = %q", entity.Name.EndTime)
	}
}
