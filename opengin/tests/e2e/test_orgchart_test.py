import requests
import json
import sys
import os
from datetime import datetime, timezone
import pandas as pd
import binascii
from google.protobuf.wrappers_pb2 import StringValue


def get_service_urls():
    read_service_url = os.getenv('READ_SERVICE_URL', 'http://0.0.0.0:8081')
    ingestion_service_url = os.getenv('INGESTION_SERVICE_URL', 'http://0.0.0.0:8080')
    
    return {
        'read': f"{read_service_url}/v1/entities",
        'ingestion': f"{ingestion_service_url}/entities"
    }

# Get service URLs from environment variables
urls = get_service_urls()
READ_API_URL = urls['read']
INGESTION_API_URL = urls['ingestion']

def format_attribute_title(attribute_name):
    """Format attribute name into a nice title with emojis and proper formatting."""
    # Convert snake_case to Title Case
    title = attribute_name.replace('_', ' ').title()
    
    # Add appropriate emojis based on attribute type
    emoji_map = {
        'Personal Information': '👤',
        'Performance Metrics': '📈',
        'Budget Allocation': '💰',
        'Department Information': '🏢',
        'Staff Information': '👥',
        'Project Portfolio': '📋',
        'Budget Breakdown': '💸'
    }
    
    emoji = emoji_map.get(title, '📊')
    return f"{emoji} {title}"


def display_tabular_data_with_pandas(data, attribute_name, show_summary=True):
    """
    Utility function to display tabular data using pandas with nice formatting.
    
    Args:
        data: The decoded tabular data (dict with 'columns' and 'rows')
        attribute_name: Name of the attribute for display purposes
        show_summary: Whether to show data summary and analysis
    
    Returns:
        pandas.DataFrame: The created DataFrame, or None if conversion failed
    """
    if not isinstance(data, dict) or 'columns' not in data or 'rows' not in data:
        print(f"    ⚠️ Unexpected data structure for {attribute_name}: {type(data)}")
        print(f"    Raw data: {data}")
        return None
    
    try:
        # Create pandas DataFrame
        df = pd.DataFrame(data['rows'], columns=data['columns'])
        
        # Format the title
        title = format_attribute_title(attribute_name)
        
        # Display the title with nice formatting
        print(f"\n    {title}")
        print("    " + "=" * (len(title) + 2))
        
        # Display basic info
        print(f"    📋 Shape: {df.shape[0]} rows × {df.shape[1]} columns")
        print(f"    📋 Columns: {list(df.columns)}")
        
        # Display the table
        print(f"\n    📊 Data Table:")
        print(df.to_string(index=False))
        
        if show_summary:
            # Show data types and basic info
            print(f"\n    📈 Data Summary:")
            print(f"    - Total Records: {len(df)}")
            print(f"    - Columns: {len(df.columns)}")
            print(f"    - Data Types:")
            for col in df.columns:
                non_null_count = df[col].count()
                print(f"      - {col}: {df[col].dtype} ({non_null_count} non-null values)")
            
            # Show intelligent analysis based on column structure
            if 'field' in df.columns and 'value' in df.columns:
                print(f"\n    🔍 Key Information:")
                for _, row in df.iterrows():
                    field = row.get('field', 'Unknown')
                    value = row.get('value', 'N/A')
                    print(f"      - {field}: {value}")
                    
            elif 'metric' in df.columns and 'target' in df.columns and 'actual' in df.columns:
                print(f"\n    📈 Performance Analysis:")
                for _, row in df.iterrows():
                    metric = row.get('metric', 'Unknown')
                    target = row.get('target', 'N/A')
                    actual = row.get('actual', 'N/A')
                    status = row.get('status', 'N/A')
                    print(f"      - {metric}: {actual}/{target} ({status})")
                    
            elif 'category' in df.columns and 'allocated_amount' in df.columns:
                print(f"\n    💰 Budget Analysis:")
                total_allocated = 0
                total_spent = 0
                for _, row in df.iterrows():
                    category = row.get('category', 'Unknown')
                    allocated = row.get('allocated_amount', 0)
                    spent = row.get('spent_amount', 0)
                    if isinstance(allocated, (int, float)):
                        total_allocated += allocated
                    if isinstance(spent, (int, float)):
                        total_spent += spent
                    print(f"      - {category}: {allocated:,} allocated, {spent:,} spent")
                print(f"      - Total: {total_allocated:,} allocated, {total_spent:,} spent")
                
            elif 'position' in df.columns and 'count' in df.columns:
                print(f"\n    👥 Staff Analysis:")
                total_staff = 0
                total_vacant = 0
                for _, row in df.iterrows():
                    position = row.get('position', 'Unknown')
                    count = row.get('count', 0)
                    vacant = row.get('vacant', 0)
                    if isinstance(count, (int, float)):
                        total_staff += count
                    if isinstance(vacant, (int, float)):
                        total_vacant += vacant
                    print(f"      - {position}: {count} total, {vacant} vacant")
                print(f"      - Total Staff: {total_staff}, Vacant: {total_vacant}")
                
            elif 'project_name' in df.columns and 'status' in df.columns:
                print(f"\n    📋 Project Analysis:")
                for _, row in df.iterrows():
                    project = row.get('project_name', 'Unknown')
                    status = row.get('status', 'N/A')
                    progress = row.get('progress', 'N/A')
                    budget = row.get('budget', 'N/A')
                    print(f"      - {project}: {status} ({progress}) - {budget}")
        
        return df
        
    except Exception as e:
        print(f"    ❌ Failed to create pandas DataFrame: {e}")
        return None


def decode_attribute_any_value(json_str: str) -> str:
    # TODO: Please check why the existing decode_protobuf_any_value is not working for attribute entities
    """
    Decode a JSON representation of a protobuf Any containing a StringValue.
    Example input:
    {"typeUrl":"type.googleapis.com/google.protobuf.StringValue",
     "value":"706572736F6E616C5F696E666F726D6174696F6E"}
    """
    data = json.loads(json_str)

    # Extract fields
    type_url = data["typeUrl"]
    hex_value = data["value"]

    if type_url != "type.googleapis.com/google.protobuf.StringValue":
        raise ValueError(f"Unsupported typeUrl: {type_url}")

    # Convert hex string -> raw bytes
    raw_bytes = binascii.unhexlify(hex_value)
    
    # The hex data appears to be the actual string content, not a protobuf message
    # Try to decode it directly as UTF-8
    try:
        return raw_bytes.decode('utf-8')
    except UnicodeDecodeError:
        # If that fails, try parsing as protobuf
        try:
            sv = StringValue()
            sv.ParseFromString(raw_bytes)
            return sv.value
        except:
            # If all else fails, return the hex value
            return hex_value


def decode_protobuf_any_value(any_value):
    """Decode a protobuf Any value to get the actual value"""
    if isinstance(any_value, dict) and 'typeUrl' in any_value and 'value' in any_value:
        type_url = any_value['typeUrl']
        value = any_value['value']

        if 'StringValue' in type_url:
            try:
                # If it's hex encoded (which appears to be the case)
                hex_value = value
                binary_data = bytes.fromhex(hex_value)
                # For StringValue in hex format, typically the structure is:
                # 0A (field tag) + 03 (length) + actual string bytes
                # Skip the first 2 bytes (field tag and length)
                if len(binary_data) > 2:
                    return binary_data[2:].decode('utf-8')
            except Exception as e:
                print(f"Failed to decode StringValue: {e}")
                return value

        elif 'Struct' in type_url:
            try:
                # For Struct type, the value is hex-encoded protobuf data
                binary_data = bytes.fromhex(value)

                # Try to use protobuf library first
                try:
                    from google.protobuf import struct_pb2
                    from google.protobuf.json_format import MessageToDict

                    struct_msg = struct_pb2.Struct()
                    struct_msg.ParseFromString(binary_data)
                    result = MessageToDict(struct_msg)

                    # If the result has a 'data' field that's a string, try to parse it as JSON
                    if isinstance(result, dict) and 'data' in result and isinstance(result['data'], str):
                        try:
                            data_json = json.loads(result['data'])
                            return data_json
                        except json.JSONDecodeError:
                            pass

                    return result

                except ImportError:
                    print("protobuf library not available, trying manual extraction")
                except Exception as e:
                    print(f"Failed to decode with protobuf library: {e}")

                # Manual extraction fallback - look for JSON patterns in the binary data
                try:
                    # Convert binary data to string and look for JSON
                    text_data = binary_data.decode('utf-8', errors='ignore')
                    print(f"Debug: Extracted text from binary: {repr(text_data[:200])}...")

                    # Find JSON-like content (look for { and })
                    start_idx = text_data.find('{')
                    end_idx = text_data.rfind('}')

                    if start_idx != -1 and end_idx != -1 and end_idx > start_idx:
                        json_str = text_data[start_idx:end_idx + 1]
                        print(f"Debug: Extracted JSON string: {repr(json_str[:100])}...")
                        return json.loads(json_str)

                except Exception as e:
                    print(f"Failed to extract JSON from binary data: {e}")

                # Try a different approach - look for specific patterns
                try:
                    # The hex data might contain the JSON in a different format
                    # Let's try to find the actual JSON content
                    text_data = binary_data.decode('utf-8', errors='ignore')

                    # Look for common JSON patterns
                    patterns = ['"columns"', '"rows"', '"data"']
                    for pattern in patterns:
                        if pattern in text_data:
                            # Find the start of the JSON object containing this pattern
                            start_idx = text_data.find('{', text_data.find(pattern) - 100)
                            if start_idx != -1:
                                # Find the matching closing brace
                                brace_count = 0
                                end_idx = start_idx
                                for i, char in enumerate(text_data[start_idx:], start_idx):
                                    if char == '{':
                                        brace_count += 1
                                    elif char == '}':
                                        brace_count -= 1
                                        if brace_count == 0:
                                            end_idx = i
                                            break

                                if end_idx > start_idx:
                                    json_str = text_data[start_idx:end_idx + 1]
                                    print(f"Debug: Found JSON with pattern {pattern}: {repr(json_str[:100])}...")
                                    try:
                                        return json.loads(json_str)
                                    except json.JSONDecodeError as e:
                                        print(f"Debug: JSON decode failed: {e}")
                                        continue

                except Exception as e:
                    print(f"Failed to extract JSON with pattern matching: {e}")

                # If all else fails, try to decode as base64
                try:
                    import base64
                    # The hex might actually be base64 encoded
                    decoded_bytes = base64.b64decode(value)
                    json_str = decoded_bytes.decode('utf-8')
                    return json.loads(json_str)
                except Exception as e:
                    print(f"Failed to decode as base64: {e}")

                # Final fallback: return the hex value
                return value

            except Exception as e:
                print(f"Failed to decode Struct: {e}")
                return value

        elif 'Int32Value' in type_url or 'Int64Value' in type_url:
            try:
                # For integer values
                hex_value = value
                binary_data = bytes.fromhex(hex_value)
                # Skip field tag and length bytes
                if len(binary_data) > 2:
                    return int.from_bytes(binary_data[2:], byteorder='little')
            except Exception as e:
                print(f"Failed to decode integer value: {e}")
                return value

        elif 'DoubleValue' in type_url or 'FloatValue' in type_url:
            try:
                # For float/double values
                hex_value = value
                binary_data = bytes.fromhex(hex_value)
                # Skip field tag and length bytes
                if len(binary_data) > 2:
                    import struct
                    return struct.unpack('<d', binary_data[2:])[0]  # little-endian double
            except Exception as e:
                print(f"Failed to decode float value: {e}")
                return value

    # If any_value is a string that looks like a JSON object
    elif isinstance(any_value, str) and any_value.startswith('{') and any_value.endswith('}'):
        try:
            # Try to parse it as JSON
            obj = json.loads(any_value)
            # Recursively decode
            return decode_protobuf_any_value(obj)
        except json.JSONDecodeError:
            pass

    # Return the original value if decoding fails
    return any_value


class OrgChartIngestion:
    """
    A class to create and manage Organizational chart data with ministers and departments.
    Creates 3 ministers, each with 2 departments, all with rich tabular attributes.
    """
    
    def __init__(self):
        self.base_url = INGESTION_API_URL
        self.created_entities = []
        
        # Define the Organizational structure
        self.ministers = [
            {
                "id": "minister-tech-001",
                "name": "Minister of Technology and Digital Innovation",
                "short_name": "Tech Minister",
                "portfolio": "Technology",
                "appointment_date": "2024-01-15T00:00:00Z"
            },
            {
                "id": "minister-health-001", 
                "name": "Minister of Health and Social Services",
                "short_name": "Health Minister",
                "portfolio": "Health",
                "appointment_date": "2024-01-15T00:00:00Z"
            },
            {
                "id": "minister-education-001",
                "name": "Minister of Education and Human Development", 
                "short_name": "Education Minister",
                "portfolio": "Education",
                "appointment_date": "2024-01-15T00:00:00Z"
            }
        ]
        
        self.departments = {
            "minister-tech-001": [
                {
                    "id": "dept-ict-001",
                    "name": "Department of Information and Communication Technology",
                    "short_name": "ICT Department",
                    "focus": "Digital Infrastructure"
                },
                {
                    "id": "dept-innovation-001", 
                    "name": "Department of Digital Innovation and Research",
                    "short_name": "Innovation Department",
                    "focus": "R&D and Innovation"
                }
            ],
            "minister-health-001": [
                {
                    "id": "dept-hospitals-001",
                    "name": "Department of Hospital Services",
                    "short_name": "Hospital Services",
                    "focus": "Healthcare Delivery"
                },
                {
                    "id": "dept-public-health-001",
                    "name": "Department of Public Health and Prevention",
                    "short_name": "Public Health",
                    "focus": "Preventive Healthcare"
                }
            ],
            "minister-education-001": [
                {
                    "id": "dept-schools-001",
                    "name": "Department of School Education",
                    "short_name": "School Education",
                    "focus": "Primary and Secondary Education"
                },
                {
                    "id": "dept-higher-ed-001",
                    "name": "Department of Higher Education and Research",
                    "short_name": "Higher Education",
                    "focus": "Universities and Research"
                }
            ]
        }
    
    def create_minister_entity(self, minister_data):
        """Create a minister entity with rich tabular attributes."""
        
        # Personal information table
        personal_info = {
            "columns": ["field", "value", "last_updated"],
            "rows": [
                ["full_name", minister_data["name"], "2024-01-15T00:00:00Z"],
                ["short_name", minister_data["short_name"], "2024-01-15T00:00:00Z"],
                ["portfolio", minister_data["portfolio"], "2024-01-15T00:00:00Z"],
                ["appointment_date", minister_data["appointment_date"], "2024-01-15T00:00:00Z"],
                ["office_location", "Parliament Complex, Colombo", "2024-01-15T00:00:00Z"],
                ["contact_email", f"{minister_data['id']}@gov.lk", "2024-01-15T00:00:00Z"],
                ["security_clearance", "Top Secret", "2024-01-15T00:00:00Z"]
            ]
        }
        
        # Performance metrics table
        performance_metrics = {
            "columns": ["metric", "target", "actual", "period", "status"],
            "rows": [
                ["budget_utilization", "95%", "92%", "Q1-2024", "On Track"],
                ["policy_implementations", "5", "3", "Q1-2024", "In Progress"],
                ["public_approval_rating", "80%", "78%", "Q1-2024", "Good"],
                ["department_efficiency", "90%", "88%", "Q1-2024", "Good"],
                ["stakeholder_satisfaction", "85%", "82%", "Q1-2024", "Good"]
            ]
        }
        
        # Budget allocation table
        budget_allocation = {
            "columns": ["category", "allocated_amount", "spent_amount", "remaining", "fiscal_year"],
            "rows": [
                ["operational_expenses", 50000000, 45000000, 5000000, "2024"],
                ["capital_investments", 200000000, 120000000, 80000000, "2024"],
                ["staff_salaries", 30000000, 30000000, 0, "2024"],
                ["research_grants", 50000000, 35000000, 15000000, "2024"],
                ["emergency_fund", 20000000, 5000000, 15000000, "2024"]
            ]
        }
        
        payload = {
            "id": minister_data["id"],
            "kind": {"major": "Organization", "minor": "Minister"},
            "created": minister_data["appointment_date"],
            "terminated": "",
            "name": {
                "startTime": minister_data["appointment_date"],
                "endTime": "",
                "value": minister_data["name"]
            },
            "metadata": [
                {"key": "portfolio", "value": minister_data["portfolio"]},
                {"key": "appointment_date", "value": minister_data["appointment_date"]},
                {"key": "entity_type", "value": "government_minister"},
                {"key": "hierarchy_level", "value": "minister"}
            ],
            "attributes": [
                {
                    "key": "personal_information",
                    "value": {
                        "values": [
                            {
                                "startTime": minister_data["appointment_date"],
                                "endTime": "",
                                "value": personal_info
                            }
                        ]
                    }
                },
                {
                    "key": "performance_metrics",
                    "value": {
                        "values": [
                            {
                                "startTime": "2024-01-01T00:00:00Z",
                                "endTime": "",
                                "value": performance_metrics
                            }
                        ]
                    }
                },
                {
                    "key": "budget_allocation",
                    "value": {
                        "values": [
                            {
                                "startTime": "2024-01-01T00:00:00Z",
                                "endTime": "2024-12-31T23:59:59Z",
                                "value": budget_allocation
                            }
                        ]
                    }
                }
            ],
            "relationships": []
        }
        
        return payload
    
    def create_department_entity(self, dept_data, minister_id):
        """Create a department entity with rich tabular attributes."""
        
        # Department information table
        dept_info = {
            "columns": ["field", "value", "last_updated"],
            "rows": [
                ["department_name", dept_data["name"], "2024-01-15T00:00:00Z"],
                ["short_name", dept_data["short_name"], "2024-01-15T00:00:00Z"],
                ["focus_area", dept_data["focus"], "2024-01-15T00:00:00Z"],
                ["establishment_date", "2024-01-15T00:00:00Z", "2024-01-15T00:00:00Z"],
                ["headquarters", "Colombo, Sri Lanka", "2024-01-15T00:00:00Z"],
                ["contact_phone", "+94-11-123-4567", "2024-01-15T00:00:00Z"],
                ["website", f"www.{dept_data['id']}.gov.lk", "2024-01-15T00:00:00Z"]
            ]
        }
        
        # Staff information table
        staff_info = {
            "columns": ["position", "count", "vacant", "salary_grade", "department"],
            "rows": [
                ["Director General", 1, 0, "SL-1", dept_data["short_name"]],
                ["Deputy Director", 2, 0, "SL-2", dept_data["short_name"]],
                ["Assistant Director", 5, 1, "SL-3", dept_data["short_name"]],
                ["Senior Officer", 15, 2, "SL-4", dept_data["short_name"]],
                ["Officer", 25, 3, "SL-5", dept_data["short_name"]],
                ["Support Staff", 30, 5, "SL-6", dept_data["short_name"]],
                ["Technical Staff", 20, 2, "SL-4", dept_data["short_name"]]
            ]
        }
        
        # Project portfolio table
        project_portfolio = {
            "columns": ["project_name", "status", "budget", "start_date", "end_date", "progress"],
            "rows": [
                ["Digital Transformation Initiative", "In Progress", 25000000, "2024-01-01", "2024-12-31", "65%"],
                ["Infrastructure Modernization", "Planning", 15000000, "2024-06-01", "2025-05-31", "10%"],
                ["Staff Training Program", "Completed", 5000000, "2024-01-01", "2024-03-31", "100%"],
                ["Policy Framework Update", "In Progress", 3000000, "2024-02-01", "2024-08-31", "40%"],
                ["Public Outreach Campaign", "Planning", 8000000, "2024-07-01", "2024-12-31", "5%"]
            ]
        }
        
        # Budget breakdown table
        budget_breakdown = {
            "columns": ["category", "allocated", "spent", "remaining", "percentage"],
            "rows": [
                ["Personnel Costs", 40000000, 38000000, 2000000, "40%"],
                ["Operational Expenses", 20000000, 15000000, 5000000, "20%"],
                ["Capital Expenditure", 30000000, 20000000, 10000000, "30%"],
                ["Programs and Projects", 10000000, 5000000, 5000000, "10%"]
            ]
        }
        
        payload = {
            "id": dept_data["id"],
            "kind": {"major": "Organization", "minor": "Department"},
            "created": "2024-01-15T00:00:00Z",
            "terminated": "",
            "name": {
                "startTime": "2024-01-15T00:00:00Z",
                "endTime": "",
                "value": dept_data["name"]
            },
            "metadata": [
                {"key": "focus_area", "value": dept_data["focus"]},
                {"key": "parent_minister", "value": minister_id},
                {"key": "entity_type", "value": "government_department"},
                {"key": "hierarchy_level", "value": "department"}
            ],
            "attributes": [
                {
                    "key": "department_information",
                    "value": {
                        "values": [
                            {
                                "startTime": "2024-01-15T00:00:00Z",
                                "endTime": "",
                                "value": dept_info
                            }
                        ]
                    }
                },
                {
                    "key": "staff_information",
                    "value": {
                        "values": [
                            {
                                "startTime": "2024-01-15T00:00:00Z",
                                "endTime": "",
                                "value": staff_info
                            }
                        ]
                    }
                },
                {
                    "key": "project_portfolio",
                    "value": {
                        "values": [
                            {
                                "startTime": "2024-01-01T00:00:00Z",
                                "endTime": "",
                                "value": project_portfolio
                            }
                        ]
                    }
                },
                {
                    "key": "budget_breakdown",
                    "value": {
                        "values": [
                            {
                                "startTime": "2024-01-01T00:00:00Z",
                                "endTime": "2024-12-31T23:59:59Z",
                                "value": budget_breakdown
                            }
                        ]
                    }
                }
            ],
            "relationships": [
                {
                    "key": "reports_to",
                    "value": {
                        "relatedEntityId": minister_id,
                        "startTime": "2024-01-15T00:00:00Z",
                        "endTime": "",
                        "id": f"rel-{dept_data['id']}-to-{minister_id}",
                        "name": "reports_to"
                    }
                }
            ]
        }
        
        return payload
    
    def create_all_entities(self):
        """Create all ministers and their departments."""
        print("\n🏛️ Creating Organizational Chart Entities...")
        
        # Create ministers first
        for minister in self.ministers:
            print(f"\n📋 Creating Minister: {minister['name']}")
            minister_payload = self.create_minister_entity(minister)
            
            response = requests.post(self.base_url, json=minister_payload)
            if response.status_code in [200, 201]:
                print(f"✅ Created Minister: {minister['id']}")
                self.created_entities.append(minister['id'])
            else:
                print(f"❌ Failed to create Minister {minister['id']}: {response.status_code} - {response.text}")
                return False
        
        # Create departments for each minister
        for minister_id, departments in self.departments.items():
            print(f"\n🏢 Creating departments for {minister_id}:")
            for dept in departments:
                print(f"  📁 Creating Department: {dept['name']}")
                dept_payload = self.create_department_entity(dept, minister_id)
                
                response = requests.post(self.base_url, json=dept_payload)
                if response.status_code in [200, 201]:
                    print(f"  ✅ Created Department: {dept['id']}")
                    self.created_entities.append(dept['id'])
                else:
                    print(f"  ❌ Failed to create Department {dept['id']}: {response.status_code} - {response.text}")
                    return False
        
        print(f"\n🎉 Successfully created {len(self.created_entities)} entities!")
        return True
    
    def get_created_entities(self):
        """Return list of created entity IDs."""
        return self.created_entities


def test_orgchart_creation():
    """Test the creation of Organizational chart entities."""
    print("\n🧪 Testing Organizational Chart Creation...")
    
    org_chart = OrgChartIngestion()
    success = org_chart.create_all_entities()
    
    if success:
        print("✅ Organizational chart creation test passed!")
        return True
    else:
        print("❌ Organizational chart creation test failed!")
        return False


def test_orgchart_query():
    """Test querying the Organizational chart data."""
    print("\n🔍 Testing Organizational Chart Queries...")
    
    # Test querying all ministers
    print("  📋 Querying all ministers...")
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization",
            "minor": "Minister"  # Only providing minor kind
        }
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search successful: {res.text}"
    
    body = res.json()
    print("📋 Ministers Query Response:")
    print(json.dumps(body, indent=2))
    
    # Validate ministers response structure - the data is in 'body' field
    assert "body" in body, "Response should contain 'body' field"
    entities = body["body"]
    assert isinstance(entities, list), "Body should be a list"
    assert len(entities) == 3, f"Expected 3 ministers, found {len(entities)}"
    
    # Validate each minister
    minister_ids = []
    for i, minister in enumerate(entities):
        print(f"\n🔍 Validating Minister {i+1}:")
        
        # Check required fields
        assert "id" in minister, f"Minister {i+1} missing 'id' field"
        assert "kind" in minister, f"Minister {i+1} missing 'kind' field"
        assert "name" in minister, f"Minister {i+1} missing 'name' field"
        
        # Validate kind structure
        kind = minister["kind"]
        assert kind["major"] == "Organization", f"Minister {i+1} has wrong major kind: {kind['major']}"
        assert kind["minor"] == "Minister", f"Minister {i+1} has wrong minor kind: {kind['minor']}"
        
        # Search returns a plain name string (not a protobuf Any blob)
        name = minister["name"]
        assert isinstance(name, str), f"Minister {i+1} name should be a string, got {type(name)}: {name}"
        assert not (name.startswith("{") and "typeUrl" in name), (
            f"Minister {i+1} name is still a serialized protobuf Any: {name}"
        )
        assert len(name) > 0, f"Minister {i+1} name should not be empty"
        print(f"  📝 Name: {name}")
        
        minister_ids.append(minister["id"])
        print(f"  ✅ Minister {i+1} validation passed: {minister['id']}")
    
    # Validate we got the expected minister IDs
    expected_minister_ids = ["minister-tech-001", "minister-health-001", "minister-education-001"]
    for expected_id in expected_minister_ids:
        assert expected_id in minister_ids, f"Expected minister {expected_id} not found in results"
    
    print(f"\n✅ All {len(entities)} ministers validated successfully!")
    print(f"📋 Minister IDs: {minister_ids}")

    # Test querying all departments
    print("  🏢 Querying all departments...")
    payload = {
        "kind": {
            "major": "Organization",
            "minor": "Department"
        }
    }
    response = requests.post(url, json=payload)
    assert response.status_code == 200, f"Search successful: {response.text}"
    
    dept_body = response.json()
    print("\n🏢 Departments Query Response:")
    print(json.dumps(dept_body, indent=2))
    
    # Validate departments response structure - the data is in 'body' field
    assert "body" in dept_body, "Response should contain 'body' field"
    dept_entities = dept_body["body"]
    assert isinstance(dept_entities, list), "Body should be a list"
    assert len(dept_entities) == 6, f"Expected 6 departments, found {len(dept_entities)}"
    
    # Validate each department
    dept_ids = []
    for i, dept in enumerate(dept_entities):
        print(f"\n🔍 Validating Department {i+1}:")
        
        # Check required fields
        assert "id" in dept, f"Department {i+1} missing 'id' field"
        assert "kind" in dept, f"Department {i+1} missing 'kind' field"
        assert "name" in dept, f"Department {i+1} missing 'name' field"
        
        # Validate kind structure
        kind = dept["kind"]
        assert kind["major"] == "Organization", f"Department {i+1} has wrong major kind: {kind['major']}"
        assert kind["minor"] == "Department", f"Department {i+1} has wrong minor kind: {kind['minor']}"
        
        # Search returns a plain name string (not a protobuf Any blob)
        name = dept["name"]
        assert isinstance(name, str), f"Department {i+1} name should be a string, got {type(name)}: {name}"
        assert not (name.startswith("{") and "typeUrl" in name), (
            f"Department {i+1} name is still a serialized protobuf Any: {name}"
        )
        assert len(name) > 0, f"Department {i+1} name should not be empty"
        print(f"  📝 Name: {name}")
        
        dept_ids.append(dept["id"])
        print(f"  ✅ Department {i+1} validation passed: {dept['id']}")
    
    # Validate we got the expected department IDs
    expected_dept_ids = [
        "dept-ict-001", "dept-innovation-001",  # Tech minister departments
        "dept-hospitals-001", "dept-public-health-001",  # Health minister departments
        "dept-schools-001", "dept-higher-ed-001"  # Education minister departments
    ]
    for expected_id in expected_dept_ids:
        assert expected_id in dept_ids, f"Expected department {expected_id} not found in results"
    
    print(f"\n✅ All {len(dept_entities)} departments validated successfully!")
    print(f"🏢 Department IDs: {dept_ids}")
    
    print("✅ Organizational chart query test passed!")
    return True


def test_get_attributes_of_one_minister():
    """Test getting all attributes for a minister by querying attribute entities."""
    print("=" * 50)
    print("\n📊 Testing Minister Attributes...")
    print("=" * 50)
    
    entity_id = "minister-tech-001"
    url = f"{READ_API_URL}/{entity_id}/relations"
    payload = {"name": "IS_ATTRIBUTE"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    print("✅ Minister relationships of type IS_ATTRIBUTE:", json.dumps(body, indent=2))
    
    # Extract relatedEntityIds from the relationships
    if isinstance(body, list):
        attribute_entity_ids = []
        for relationship in body:
            if 'relatedEntityId' in relationship:
                attribute_entity_ids.append(relationship['relatedEntityId'])
                print(f"  🔍 Found attribute entity: {relationship['relatedEntityId']}")
        
        print(f"\n📋 Found {len(attribute_entity_ids)} attribute entities")
        
        # Query each attribute entity using search endpoint
        for attr_entity_id in attribute_entity_ids:
            print(f"\n  🔍 Querying attribute entity: {attr_entity_id}")
            search_url = f"{READ_API_URL}/search"
            search_payload = {"id": attr_entity_id}
            
            search_res = requests.post(search_url, json=search_payload)
            if search_res.status_code == 200:
                attr_body = search_res.json()["body"]

                print(f"  ✅ Successfully retrieved attribute entity data")
                
                # Just print what's in the object
                print(f"  📊 Attribute entity data: {json.dumps(attr_body, indent=2)}")
                name = attr_body[0]["name"]
                assert isinstance(name, str), f"Attribute name should be a string, got {type(name)}: {name}"
                print(f"  ✅ Attribute Name: {name}")

            else:
                print(f"  ❌ Failed to query attribute entity {attr_entity_id}: {search_res.status_code}")
                return False
    else:
        print("  ⚠️ Unexpected response structure for relationships")
        return False
    
    return True


def test_one_minister_attributes_with_pandas():
    """Test getting all attributes for a minister and display them with pandas."""
    print("=" * 50)
    print("\n📊 Testing Minister Attributes with Pandas...")
    print("=" * 50)

    entity_id = "minister-tech-001"
    attributes_to_test = [
        "personal_information",
        "performance_metrics", 
        "budget_allocation"
    ]
    
    for attr_name in attributes_to_test:
        print(f"\n  🔍 Testing {attr_name}...")
        url = f"{READ_API_URL}/{entity_id}/attributes/{attr_name}"
        
        res = requests.get(url)
        if res.status_code == 200:
            data = res.json()
            start_time = data['start']
            end_time = data['end']
            value = data['value']
            decoded_value = decode_protobuf_any_value(value)
            
            print(f"    📅 Time Range: {start_time} to {end_time}")
            
            # Use the utility function to display data with pandas
            df = display_tabular_data_with_pandas(decoded_value, attr_name, show_summary=True)
        else:
            print(f"    ❌ Failed to query {attr_name}: {res.status_code}")
            return False
    
    print(f"\n✅ All attributes displayed successfully with pandas!")
    return True


def test_one_department_attributes_with_pandas():
    """Test getting all attributes for a minister and display them with pandas."""
    print("=" * 50)
    print("\n📊 Testing Department Attributes with Pandas...")
    print("=" * 50)
    
    entity_id = "dept-ict-001"
    attributes_to_test = [
        "department_information",
        "staff_information", 
        "project_portfolio",
        "budget_breakdown",
    ]
    
    for attr_name in attributes_to_test:
        print(f"\n  🔍 Testing {attr_name}...")
        url = f"{READ_API_URL}/{entity_id}/attributes/{attr_name}"
        
        res = requests.get(url)
        if res.status_code == 200:
            data = res.json()
            start_time = data['start']
            end_time = data['end']
            value = data['value']
            decoded_value = decode_protobuf_any_value(value)
            
            print(f"    📅 Time Range: {start_time} to {end_time}")
            
            # Use the utility function to display data with pandas
            df = display_tabular_data_with_pandas(decoded_value, attr_name, show_summary=True)
        else:
            print(f"    ❌ Failed to query {attr_name}: {res.status_code}")
            return False
    
    print(f"\n✅ All attributes displayed successfully with pandas!")
    return True


if __name__ == "__main__":
    print("🏛️ Organizational Chart Test Suite")
    print("=" * 50)
    
    # Test creation
    creation_success = test_orgchart_creation()
    
    if creation_success:
        # Test basic querying
        query_success = test_orgchart_query()

        if query_success:

            relations_success = test_get_attributes_of_one_minister()
            
            if relations_success:
                print("\n🎉 All relatons !")
            else:
                print("\n❌ Minister relationships of type IS_ATTRIBUTE failed!")
        
            if relations_success:
                # Test all attributes with pandas display
                minister_attributes_success = test_one_minister_attributes_with_pandas()
                
                if minister_attributes_success:
                    print("\n🎉 All organizational chart tests with pandas passed!")
                    print("=" * 50)
                    print("📊 Test Summary:")
                    print("  ✅ Entity Creation: 3 Ministers + 6 Departments")
                    print("  ✅ Basic Queries: Minister and Department queries")
                    print("  ✅ Attribute Queries: Individual attribute retrieval")
                    print("  ✅ Pandas Display: Tabular data visualization and analysis")
                else:
                    print("\n❌ Minister attributes display tests failed!")

                department_attributes_success = test_one_department_attributes_with_pandas()
                
                if department_attributes_success:
                    print("\n🎉 All organizational chart tests with pandas passed!")
                    print("=" * 50)
                    print("📊 Test Summary:")
                    print("  ✅ Entity Creation: 3 Ministers + 6 Departments")
                    print("  ✅ Basic Queries: Minister and Department queries")
                    print("  ✅ Attribute Queries: Individual attribute retrieval")
                    print("  ✅ Pandas Display: Tabular data visualization and analysis")
                    sys.exit(0)
                else:
                    print("\n❌ Department attributes display tests failed!")
            else:
                print("\n❌ Basic query tests failed!")
        else:
            print("\n❌ Relations tests failed!")

    else:
        print("\n❌ Creation tests failed!")
        sys.exit(1)
