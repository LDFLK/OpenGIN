import pandas as pd
import requests
import json
import os
import argparse
import sys
from datetime import datetime

# Configuration
# OpenGIN uses separate ports for Ingestion (writes) and Reading
INGESTION_API = "http://localhost:8080/entities"  # gRPC-Gateway for ingestion
READ_API = "http://localhost:8081/v1/entities"   # REST API for reading
DATA_DIR = "tutorials/data"

def load_data():
    """Load all data files from the data directory."""
    print("📂 Loading data files...")
    
    entities_df = pd.read_csv(os.path.join(DATA_DIR, "entities.csv"))
    relationships_df = pd.read_csv(os.path.join(DATA_DIR, "relationships.csv"))
    
    # Load attributes if they exist
    try:
        expenses_df = pd.read_csv(os.path.join(DATA_DIR, "expenses.csv"))
    except FileNotFoundError:
        expenses_df = pd.DataFrame()
        
    try:
        income_df = pd.read_csv(os.path.join(DATA_DIR, "income.csv"))
    except FileNotFoundError:
        income_df = pd.DataFrame()

    # Load metadata
    try:
        with open(os.path.join(DATA_DIR, "metadata.json"), "r") as f:
            metadata_dict = json.load(f)
    except FileNotFoundError:
        metadata_dict = {}

    print(f"✅ Loaded {len(entities_df)} entities, {len(relationships_df)} relationships")
    return entities_df, relationships_df, expenses_df, income_df, metadata_dict

def create_tabular_attribute(key, df, relevant_columns):
    """
    Convert a DataFrame subset into OpenGIN's 'tabular' attribute format.
    
    In OpenGIN, attributes can represent complex structures like tables.
    A tabular attribute consists of:
    1. 'columns': List of column headers.
    2. 'rows': List of rows, where each row is a list of values.
    
    This function wraps the table in a TimeBasedValue, as all attributes 
    in OpenGIN are versioned by time.
    """
    if df.empty:
        return None
        
    # Prepare columns and rows
    columns = relevant_columns
    rows = df[columns].astype(str).values.tolist()
    
    # The actual data structure relative to the attribute type
    table_value = {
        "columns": columns,
        "rows": rows
    }
    
    # Wrap in TimeBasedValue structure
    # startTime indicates when this version of the data became valid
    return {
        "key": key,
        "value": {
            "values": [
                {
                    "startTime": datetime.now().isoformat() + "Z",
                    "value": table_value
                }
            ]
        }
    }

def construct_payload(entity_id, entity, relationships, expenses, income, metadata, include_relationships=False, exclude_kind=False):
    """Construct the entity payload."""
    
    # 1. Base Structure
    payload = {
        "id": entity_id,
        "created": entity["start_time"],
        "name": {
            "startTime": entity["start_time"],
            "value": entity["name"]
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    }

    if not exclude_kind:
        payload["kind"] = {
            "major": entity["kind_major"],
            "minor": entity["kind_minor"]
        }
    
    # 2. Metadata
    if entity_id in metadata:
        for k, v in metadata[entity_id].items():
            payload["metadata"].append({"key": k, "value": v})
            
    # 3. Attributes (Expenses)
    entity_expenses = expenses[expenses["RelatedEntityID"] == entity_id] if not expenses.empty and "RelatedEntityID" in expenses.columns else pd.DataFrame()
    if not entity_expenses.empty:
        attr = create_tabular_attribute("expenses", entity_expenses, ["Month", "Category", "Amount"])
        if attr:
            payload["attributes"].append(attr)
            
    # 4. Attributes (Income)
    entity_income = income[income["RelatedEntityID"] == entity_id] if not income.empty and "RelatedEntityID" in income.columns else pd.DataFrame()
    if not entity_income.empty:
        attr = create_tabular_attribute("income", entity_income, ["Month", "Source", "Amount"])
        if attr:
            payload["attributes"].append(attr)
            
    # 5. Relationships
    # NOTE: Relationships are processed but not always included in the initial creation.
    # See ingest_data() for the 2-pass strategy explanation.
    if include_relationships:
        entity_rels = relationships[relationships["from_id"] == entity_id]
        for _, rel in entity_rels.iterrows():
            rel_name = rel["relation"]
            # IMPORTANT: OpenGIN requires a unique ID for every relationship instance.
            rel_id = f"rel-{entity_id}-to-{rel['to_id']}".lower() 
            
            rel_payload = {
                "key": rel_name, 
                "value": {
                    "relatedEntityId": rel["to_id"],
                    # Relationships are also TimeBasedValues with start/end times
                    "startTime": rel["start_time"],
                    "endTime": "", 
                    "id": rel_id,
                    "name": rel_name
                }
            }
            payload["relationships"].append(rel_payload)
            
    return payload

def send_request(method, url, payload):
    """Send HTTP request and print status."""
    try:
        if method == "POST":
            response = requests.post(url, json=payload)
        elif method == "PUT":
            response = requests.put(url, json=payload)
        
        if response.status_code in [200, 201]:
            print(f"    ✅ Success")
            return True
        else:
            # Handle specific error cases gracefully
            err_msg = response.text
            if "already exists" in err_msg:
                 print(f"    ⚠️  Skipping creation (Entity already exists)")
                 return True
            elif "cannot update immutable fields" in err_msg:
                 print(f"    ⚠️  Skipping relationship update (Already linked)")
                 return True
            
            print(f"    ❌ Failed: {response.status_code} - {err_msg}")
            # print(f"    Payload was: {json.dumps(payload, indent=2)}") 
            return False
    except Exception as e:
        print(f"    ❌ Error: {e}")
        return False



def ingest_data():
    """Process and ingest data into OpenGIN in two phases."""
    
    entities, relationships, expenses, income, metadata = load_data()
    
    print("\n🚀 Phase 1: Creating Entities (without relationships)...")
    
    # Phase 1: Create Entities
    for _, entity in entities.iterrows():
        entity_id = entity["id"]
        print(f"  Creating Entity: {entity_id} ({entity['name']})...")
        
        payload = construct_payload(entity_id, entity, relationships, expenses, income, metadata, include_relationships=False)
        send_request("POST", INGESTION_API, payload)

    print("\n🔗 Phase 2: Updating Entities (with relationships)...")
    
    # -------------------------------------------------------------------------
    # CRITICAL: Relationship Creation Strategy
    # -------------------------------------------------------------------------
    # In OpenGIN, creating a relationship (graph edge) requires both the
    # Source and Destination entities to exist in the database to satisfy
    # referential integrity.
    #
    # Valid:   Entity A (exists) --[REL]--> Entity B (exists)
    # Invalid: Entity A (creating) --[REL]--> Entity B (does not exist yet)
    #
    # Therefore, we use a 2-Pass Ingestion Strategy:
    # 1. CREATE all entities first (without relationships).
    # 2. UPDATE entities to add relationships once we know all targets exist.
    # -------------------------------------------------------------------------
    
    # Only process entities that actually have relationships originating from them
    entities_with_rels = relationships["from_id"].unique()
    
    for _, entity in entities.iterrows():
        entity_id = entity["id"]
        
        if entity_id in entities_with_rels:
            print(f"  Linking Entity: {entity_id}...")
            
            # Re-construct payload WITH relationships, but WITHOUT kind
            payload = construct_payload(entity_id, entity, relationships, expenses, income, metadata, include_relationships=True, exclude_kind=True)
            
            # Use PUT to update the entity
            url = f"{INGESTION_API}/{entity_id}"
            send_request("PUT", url, payload)

def verify_ingestion():
    """Verify data by querying the Read API."""
    print("\n🔍 Verifying Ingestion...")
    # Clean verification logic: Search for 'Person' major kind
    search_url = f"{READ_API}/search"
    query = {
        "kind": {
            "major": "Person"
        }
    }
    
    try:
        response = requests.post(search_url, json=query)
        if response.status_code == 200:
            results = response.json().get("body", [])
            print(f"✅ Found {len(results)} Person entities.")
            for r in results:
                print(f"   - {r.get('id')} ({r.get('kind', {}).get('minor')})")
        else:
             print(f"⚠️ Search failed or empty: {response.status_code}")
    except Exception as e:
        print(f"⚠️ Verification failed: {e}")

def main():
    parser = argparse.ArgumentParser(description="OpenGIN Tutorial Application CLI")
    subparsers = parser.add_subparsers(dest="command", help="Available commands")

    # Command: create
    parser_create = subparsers.add_parser("create", help="Ingest data into OpenGIN")
    
    # Command: verify
    parser_verify = subparsers.add_parser("verify", help="Verify ingestion results")

    args = parser.parse_args()

    if args.command == "create":
        ingest_data()
        verify_ingestion()

    elif args.command == "verify":
        verify_ingestion()
    else:
        parser.print_help()
        sys.exit(1)

if __name__ == "__main__":
    main()
