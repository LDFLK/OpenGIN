from neo4j import GraphDatabase
import sys
import os

# Configuration
# Default to localhost for local testing against forwarded ports or mapped ports
URI = os.getenv("NEO4J_URI", "bolt://localhost:7687")
USERNAME = os.getenv("NEO4J_USER", "neo4j")
PASSWORD = os.getenv("NEO4J_PASSWORD", "neo4j123")

def check_connectivity(driver):
    print(f"1. Checking connectivity to {URI}...")
    try:
        driver.verify_connectivity()
        print("   include connectivity check: PASSED")
    except Exception as e:
        print(f"   include connectivity check: FAILED - {e}")
        raise e

def count_nodes(driver):
    print("2. Checking total node count...")
    query = "MATCH (n) RETURN count(n) as count"
    with driver.session() as session:
        result = session.run(query)
        record = result.single()
        count = record["count"]
        print(f"   Total nodes found: {count}")
        if count == 0:
            print("   WARNING: Database appears empty.")
        else:
             print("   Node count check: PASSED")

def verify_specific_node(driver):
    print("3. Verifying specific data...")
    query = 'MATCH (n:Organisation {Id: "2153-12_dep_129"}) RETURN n LIMIT 25'
    
    with driver.session() as session:
        print(f"   Running query: {query}")
        result = session.run(query)
        record = result.single()
        
        if record:
            node = record["n"]
            name = node.get("Name")
            print(f"   Found Node with Name: {name}")
            
            if name == "Council of Legal Education":
                print("   Specific data verification: PASSED")
            else:
                raise Exception(f"Node found but Name mismatch. Expected 'Council of Legal Education', got '{name}'")
        else:
            raise Exception("No node found with Id '2153-12_dep_129'")

def main():
    driver = None
    try:
        driver = GraphDatabase.driver(URI, auth=(USERNAME, PASSWORD))
        
        check_connectivity(driver)
        count_nodes(driver)
        verify_specific_node(driver)
        
        print("\nAll verification checks completed successfully.")
        sys.exit(0)
        
    except Exception as e:
        print(f"\nVerification FAILED: {e}")
        sys.exit(1)
    finally:
        if driver:
            driver.close()

if __name__ == "__main__":
    main()
