// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"fmt"
	pb "lk/datafoundation/core-api/lk/datafoundation/core-api"
	"log"
)

func debugMetadata(req *pb.Entity) {

	log.Printf("Reading Entity metadata: %s", req.Id)
	fmt.Println("DEBUG - Received request:")
	fmt.Printf("DEBUG - Entity ID: %s\n", req.Id)
	fmt.Printf("DEBUG - Kind: %+v\n", req.Kind)
	fmt.Printf("DEBUG - Created: %s\n", req.Created)
	fmt.Printf("DEBUG - Terminated: %s\n", req.Terminated)
	fmt.Printf("DEBUG - Name: %+v\n", req.Name)

	// Print all metadata entries
	fmt.Println("DEBUG - Metadata entries count:", len(req.Metadata))
	for key, value := range req.Metadata {
		fmt.Printf("DEBUG - Metadata key: %s\n", key)
		if value != nil {
			fmt.Printf("DEBUG - Metadata value typeUrl: %s\n", value.TypeUrl)
			fmt.Printf("DEBUG - Metadata value data length: %d\n", len(value.Value))
		} else {
			fmt.Println("DEBUG - Metadata value is nil")
		}
	}

}

func debugUtils(req *pb.Entity) {

	// Print attributes if present
	fmt.Println("DEBUG - Attributes count:", len(req.Attributes))
	for key, valueList := range req.Attributes {
		fmt.Printf("DEBUG - Attribute key: %s\n", key)
		if valueList != nil {
			fmt.Printf("DEBUG - Attribute values count: %d\n", len(valueList.Values))
		}
	}

	// Print relationships if present
	fmt.Println("DEBUG - Relationships count:", len(req.Relationships))
	for key, rel := range req.Relationships {
		fmt.Printf("DEBUG - Relationship key: %s\n", key)
		if rel != nil {
			fmt.Printf("DEBUG - Related entity ID: %s\n", rel.RelatedEntityId)
		}
	}
}
