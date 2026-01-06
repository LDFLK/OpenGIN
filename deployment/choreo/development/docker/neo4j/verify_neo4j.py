from neo4j import GraphDatabase
import sys
import os

# Configuration
# Default to localhost for local testing against forwarded ports or mapped ports
URI = os.getenv("NEO4J_URI", "bolt://localhost:7687")
USERNAME = os.getenv("NEO4J_USER", "neo4j")
PASSWORD = os.getenv("NEO4J_PASSWORD", "neo4j123")

def verify_data():
    print(f"Connecting to {URI} as {USERNAME}...")
    try:
        driver = GraphDatabase.driver(URI, auth=(USERNAME, PASSWORD))
        driver.verify_connectivity()
    except Exception as e:
        print(f"Failed to connect to Neo4j: {e}")
        sys.exit(1)

    query = 'MATCH (n:Organisation {Id: "2153-12_dep_129"}) RETURN n LIMIT 25'
    
    try:
        with driver.session() as session:
            print(f"Running query: {query}")
            result = session.run(query)
            record = result.single()
            
            if record:
                node = record["n"]
                name = node.get("Name")
                print(f"Found Node with Name: {name}")
                
                if name == "Council of Legal Education":
                    print("SUCCESS: Verification Passed! 'Council of Legal Education' found.")
                    sys.exit(0)
                else:
                    print(f"FAILURE: Node found but Name mismatch. Expected 'Council of Legal Education', got '{name}'")
                    sys.exit(1)
            else:
                print("FAILURE: No node found with Id '2153-12_dep_129'")
                sys.exit(1)
                
    except Exception as e:
        print(f"An error occurred during verification: {e}")
        sys.exit(1)
    finally:
        driver.close()

if __name__ == "__main__":
    verify_data()
