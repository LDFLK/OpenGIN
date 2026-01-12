import requests
import json
import sys
import os

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

ENTITY_ID = "read-test-entity"
RELATED_ID_1 = "read-related-entity-1"
RELATED_ID_2 = "read-related-entity-2"
RELATED_ID_3 = "read-related-entity-3"
RELATED_ID_4 = "read-related-entity-4"

# Constants for government organization test
GOVERNMENT_ID = "gov-lk-001"
MINISTER_ID_1 = "minister-tech-001"
MINISTER_ID_2 = "minister-health-001"
DEPT_ID_1 = "dept-it-001"
DEPT_ID_2 = "dept-digital-001"
DEPT_ID_3 = "dept-hospitals-001"
DEPT_ID_4 = "dept-pharma-001"

"""
The current tests only contain metadata validation.
"""

def validate_tabular_data_structure(data, expected_columns=None, min_rows=0, max_rows=None):
    """
    Validate that data has the expected tabular structure.

    Args:
        data: The data to validate
        expected_columns: List of expected column names (optional)
        min_rows: Minimum number of rows expected
        max_rows: Maximum number of rows expected (optional)

    Returns:
        dict: Validation result with 'valid' boolean and 'message' string
    """
    if not isinstance(data, dict):
        return {"valid": False, "message": f"Expected dict, got {type(data)}"}

    if 'columns' not in data:
        return {"valid": False, "message": "Missing 'columns' field"}

    if 'rows' not in data:
        return {"valid": False, "message": "Missing 'rows' field"}

    columns = data['columns']
    rows = data['rows']

    if not isinstance(columns, list):
        return {"valid": False, "message": f"'columns' should be a list, got {type(columns)}"}

    if not isinstance(rows, list):
        return {"valid": False, "message": f"'rows' should be a list, got {type(rows)}"}

    if len(rows) < min_rows:
        return {"valid": False, "message": f"Expected at least {min_rows} rows, got {len(rows)}"}

    if max_rows is not None and len(rows) > max_rows:
        return {"valid": False, "message": f"Expected at most {max_rows} rows, got {len(rows)}"}

    if expected_columns is not None:
        if set(columns) != set(expected_columns):
            return {"valid": False, "message": f"Expected columns {expected_columns}, got {columns}"}

    # Validate that each row has the same number of columns
    for i, row in enumerate(rows):
        if not isinstance(row, list):
            return {"valid": False, "message": f"Row {i} should be a list, got {type(row)}"}
        if len(row) != len(columns):
            return {"valid": False, "message": f"Row {i} has {len(row)} values but expected {len(columns)} columns"}

    return {"valid": True, "message": "Tabular data structure is valid"}

def validate_tabular_data_content(data, expected_data=None, field_filter=None):
    """
    Validate the content of tabular data.

    Args:
        data: The tabular data to validate
        expected_data: Expected tabular data structure (optional)
        field_filter: List of fields to validate (optional)

    Returns:
        dict: Validation result with 'valid' boolean and 'message' string
    """
    structure_validation = validate_tabular_data_structure(data)
    if not structure_validation['valid']:
        return structure_validation

    columns = data['columns']
    rows = data['rows']

    # If field filter is provided, validate only those fields
    if field_filter is not None:
        if not all(field in columns for field in field_filter):
            missing_fields = [f for f in field_filter if f not in columns]
            return {"valid": False, "message": f"Missing expected fields: {missing_fields}"}

        # Filter columns and rows to only include requested fields
        field_indices = [columns.index(field) for field in field_filter]
        filtered_columns = [columns[i] for i in field_indices]
        filtered_rows = [[row[i] for i in field_indices] for row in rows]

        columns = filtered_columns
        rows = filtered_rows

    # If expected data is provided, validate against it
    if expected_data is not None:
        expected_validation = validate_tabular_data_structure(expected_data)
        if not expected_validation['valid']:
            return {"valid": False, "message": f"Invalid expected data: {expected_validation['message']}"}

        expected_columns = expected_data['columns']
        expected_rows = expected_data['rows']

        if columns != expected_columns:
            return {"valid": False, "message": f"Column mismatch: expected {expected_columns}, got {columns}"}

        if len(rows) != len(expected_rows):
            return {"valid": False, "message": f"Row count mismatch: expected {len(expected_rows)}, got {len(rows)}"}

        # Validate each row
        for i, (actual_row, expected_row) in enumerate(zip(rows, expected_rows)):
            if actual_row != expected_row:
                return {"valid": False, "message": f"Row {i} mismatch: expected {expected_row}, got {actual_row}"}

    return {"valid": True, "message": "Tabular data content is valid"}

def assert_tabular_data(data, expected_columns=None, expected_data=None, field_filter=None, min_rows=0, max_rows=None):
    """
    Assert that tabular data meets the expected criteria.

    Args:
        data: The data to validate
        expected_columns: List of expected column names (optional)
        expected_data: Expected tabular data structure (optional)
        field_filter: List of fields to validate (optional)
        min_rows: Minimum number of rows expected
        max_rows: Maximum number of rows expected (optional)

    Raises:
        AssertionError: If validation fails
    """
    validation = validate_tabular_data_content(data, expected_data, field_filter)

    if not validation['valid']:
        raise AssertionError(f"Tabular data validation failed: {validation['message']}")

    # Additional structure validation
    structure_validation = validate_tabular_data_structure(data, expected_columns, min_rows, max_rows)
    if not structure_validation['valid']:
        raise AssertionError(f"Tabular data structure validation failed: {structure_validation['message']}")

def test_api_endpoint_with_validation(url, params=None, expected_fields=None, min_rows=0, test_name="API Test"):
    """
    Generic function to test any API endpoint with tabular data validation.

    Args:
        url: The API endpoint URL
        params: Query parameters (optional)
        expected_fields: Expected field names (optional)
        min_rows: Minimum number of rows expected
        test_name: Name for the test (for logging)

    Raises:
        AssertionError: If validation fails
    """
    print(f"  📋 {test_name}...")

    res = requests.get(url, params=params)

    assert res.status_code == 200, f"HTTP {res.status_code}: {res.text}"

    data = res.json()

    # Decode protobuf value if present
    assert 'value' in data, f"{test_name} - No 'value' field in response"
    
    value = data['value']
    decoded_data = decode_protobuf_any_value(value)
    print(f"    ✅ Decoded data: {decoded_data}")

    # Validate tabular data - raises AssertionError if validation fails
    assert_tabular_data(decoded_data,
                      expected_columns=expected_fields,
                      field_filter=expected_fields,
                      min_rows=min_rows)

    print(f"    ✅ {test_name} passed validation")

def get_expected_employee_data():
    """Get the expected employee data structure for validation."""
    return {
        "columns": ["id", "name", "age", "department", "salary"],
        "rows": [
            [1, "John Doe", 30, "Engineering", 75000.5],
            [2, "Jane Smith", 25, "Marketing", 65000],
            [3, "Bob Wilson", 35, "Sales", 85000.75],
            [4, "Alice Brown", 28, "Engineering", 70000.25],
            [5, "Charlie Davis", 32, "Finance", 80000]
        ]
    }

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

                    # Look for the actual start of JSON by searching for common patterns
                    # like {"columns" or {"rows" which indicate the start of our data
                    json_patterns = ['{"columns"', '{"rows"', '{"data"']
                    start_idx = -1
                    for pattern in json_patterns:
                        idx = text_data.find(pattern)
                        if idx != -1:
                            start_idx = idx
                            break
                    
                    if start_idx != -1:
                        # Find the matching closing brace
                        end_idx = text_data.rfind('}')
                        if end_idx != -1 and end_idx > start_idx:
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

def test_generic_validation_examples():
    """Test examples using the generic validation function."""
    print("\n🧪 Testing generic validation examples...")

    base_url = f"{READ_API_URL}/{ENTITY_ID}/attributes/employee_data"

    # Example 1: Test all fields
    test_api_endpoint_with_validation(
        url=base_url,
        params={"fields": ["id", "name", "age", "department", "salary"]},
        expected_fields=["id", "name", "age", "department", "salary"],
        min_rows=5,
        test_name="All Fields Test"
    )

    # Example 2: Test specific fields
    test_api_endpoint_with_validation(
        url=base_url,
        params={"fields": ["id", "name"]},
        expected_fields=["id", "name"],
        min_rows=5,
        test_name="ID and Name Fields Test"
    )

    # Example 3: Test with time range
    test_api_endpoint_with_validation(
        url=base_url,
        params={"startTime": "2024-01-01T00:00:00Z", "fields": ["salary", "department"]},
        expected_fields=["salary", "department"],
        min_rows=5,
        test_name="Time Range with Salary/Department Test"
    )

    print("  ✅ Generic validation examples completed")

def test_comprehensive_validation():
    """Test comprehensive validation of tabular data responses."""
    print("\n🧪 Testing comprehensive validation...")

    # Test the validation functions with sample data
    print("  📋 Testing validation functions...")

    # Valid tabular data
    valid_data = {
        "columns": ["id", "name", "salary"],
        "rows": [
            [1, "John Doe", 75000.5],
            [2, "Jane Smith", 65000]
        ]
    }

    # Test structure validation
    result = validate_tabular_data_structure(valid_data, expected_columns=["id", "name", "salary"], min_rows=2)
    assert result['valid'], f"Structure validation should pass: {result['message']}"
    print("    ✅ Structure validation passed")

    # Test content validation
    result = validate_tabular_data_content(valid_data, field_filter=["id", "name"])
    assert result['valid'], f"Content validation should pass: {result['message']}"
    print("    ✅ Content validation passed")

    # Test assertion function
    assert_tabular_data(valid_data, 
                      expected_columns=["id", "name", "salary"],
                      field_filter=["id", "name"],
                      min_rows=2)
    print("    ✅ Assertion function passed")

    # Test invalid data
    invalid_data = {
        "columns": ["id", "name"],
        "rows": [
            [1, "John Doe", 75000.5]  # Wrong number of values
        ]
    }

    result = validate_tabular_data_structure(invalid_data)
    assert not result['valid'], "Should detect invalid structure"
    print("    ✅ Invalid data detection passed")

    print("  ✅ Comprehensive validation tests completed")

def test_protobuf_decoding():
    """Test the protobuf decoding function with sample data"""
    print("\n🧪 Testing protobuf decoding...")

    # Test data similar to what you're getting
    test_data = {
        "typeUrl": "type.googleapis.com/google.protobuf.Struct",
        "value": "0AB4010A046461746112AB011AA8017B22636F6C756D6E73223A5B226964222C226E616D65222C2273616C617279225D2C22726F7773223A5B5B312C224A6F686E20446F65222C37353030302E355D2C5B322C224A616E6520536D697468222C36353030305D2C5B332C22426F622057696C736F6E222C38353030302E37355D2C5B342C22416C6963652042726F776E222C37303030302E32355D2C5B352C22436861726C6965204461766973222C38303030305D5D7D"
    }

    print("  📋 Testing Struct decoding...")

    # First, let's examine the hex data directly
    hex_value = test_data['value']
    print(f"    🔍 Hex value length: {len(hex_value)}")
    print(f"    🔍 First 100 chars: {hex_value[:100]}")

    # Convert to binary and examine
    try:
        binary_data = bytes.fromhex(hex_value)
        print(f"    🔍 Binary data length: {len(binary_data)}")

        # Try to decode as UTF-8
        text_data = binary_data.decode('utf-8', errors='ignore')
        print(f"    🔍 Decoded text (first 200 chars): {repr(text_data[:200])}")

        # Look for JSON patterns
        if '"columns"' in text_data:
            print("    ✅ Found 'columns' in text data")
        if '"rows"' in text_data:
            print("    ✅ Found 'rows' in text data")
        if '"data"' in text_data:
            print("    ✅ Found 'data' in text data")

    except Exception as e:
        print(f"    ❌ Failed to examine hex data: {e}")

    # Now try the actual decoding
    decoded = decode_protobuf_any_value(test_data)
    print(f"    ✅ Decoded result type: {type(decoded)}")
    print(f"    ✅ Decoded result: {decoded}")

    # Check if we got the expected tabular data structure
    if isinstance(decoded, dict):
        if 'columns' in decoded and 'rows' in decoded:
            print(f"    📊 Tabular data found!")
            print(f"    📊 Columns: {decoded['columns']}")
            print(f"    📊 Rows: {decoded['rows']}")
        elif 'data' in decoded:
            print(f"    📊 Data field: {decoded['data']}")
            if isinstance(decoded['data'], str):
                try:
                    data_json = json.loads(decoded['data'])
                    print(f"    📊 Parsed JSON data: {data_json}")
                except json.JSONDecodeError as e:
                    print(f"    ❌ Failed to parse data as JSON: {e}")
        else:
            print(f"    📊 Other dict structure: {list(decoded.keys())}")
    else:
        print(f"    📊 Non-dict result: {decoded}")

def create_entity_for_read():
    """Create a base entity with metadata, attributes, and relationships."""
    print("\n🟢 Creating entity for read tests...")

# First related entity
    payload_child_1 = {
        "id": RELATED_ID_1,
        "kind": {"major": "test", "minor": "child"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Read Test Entity Child 1"
        },
        "metadata": [
            {"key": "source", "value": "unit-test-1"},
            {"key": "env", "value": "test-1"}
        ],
        "attributes": [
            {
                "key": "employee_data",
                "value": {
                    "values": [
                        {
                            "startTime": "2024-11-01T00:00:00Z",
                            "endTime": "",
                            "value": {
                                "columns": ["e_id", "name", "age", "department", "salary"],
                                "rows": [
                                    [1, "John Doe", 30, "Engineering", 75000.50],
                                    [2, "Jane Smith", 25, "Marketing", 65000],
                                    [3, "Bob Wilson", 35, "Sales", 85000.75],
                                    [4, "Alice Brown", 28, "Engineering", 70000.25],
                                    [5, "Charlie Davis", 32, "Finance", 80000]
                                ]
                            }
                        }
                    ]
                }
            }
        ],
        "relationships": [
        ]
    }

    # FIXME: https://github.com/LDFLK/nexoan/issues/235
    # TODO: note that the attribute humidity is a scalar value and this must be saved 
    #  as a scalar value and it should be handled as a Document type. Single key value pair. 
    #  The current implementation only supports saving tabular data.

    # Second related entity
    payload_child_2 = {
        "id": RELATED_ID_2,
        "kind": {"major": "test", "minor": "child"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Read Test Entity Child 2"
        },
        "metadata": [
            {"key": "source", "value": "unit-test-2"},
            {"key": "env", "value": "test-2"}
        ],
        "attributes": [],
        "relationships": []
    }

    # Third related entity
    
    payload_child_3 = {
        "id": RELATED_ID_3,
        "kind": {"major": "test", "minor": "child"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Read Test Entity Child 3"
        },
        "metadata": [
            {"key": "source", "value": "unit-test-3"},
            {"key": "env", "value": "test-3"}
        ],
        "attributes": [],
        "relationships": []
    }

    payload_source = {
        "id": ENTITY_ID,
        "kind": {"major": "test", "minor": "parent"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Read Test Entity"
        },
        "metadata": [
            {"key": "source", "value": "unit-test"},
            {"key": "env", "value": "test"}
        ],
        "attributes": [
            {
                "key": "employee_data",
                "value": {
                    "values": [
                        {
                            "startTime": "2024-11-01T00:00:00Z",
                            "endTime": "",
                            "value": {
                                "columns": ["e_id", "name", "age", "department", "salary"],
                                "rows": [
                                    [1, "John Doe", 30, "Engineering", 75000.50],
                                    [2, "Jane Smith", 25, "Marketing", 65000],
                                    [3, "Bob Wilson", 35, "Sales", 85000.75],
                                    [4, "Alice Brown", 28, "Engineering", 70000.25],
                                    [5, "Charlie Davis", 32, "Finance", 80000]
                                ]
                            }
                        }
                    ]
                }
            }
        ],
        "relationships": [
            {
                "key": "rel-001",
                "value": {
                    "relatedEntityId": RELATED_ID_1,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "2024-12-31T23:59:59Z",
                    "id": "rel-001",
                    "name": "linked"
                }
            },
            {
                "key": "rel-002",
                "value": {
                    "relatedEntityId": RELATED_ID_2,
                    "startTime": "2024-06-01T00:00:00Z",  # Different timestamp
                    "endTime": "2024-12-31T23:59:59Z",
                    "id": "rel-002",
                    "name": "linked"  # Same type as the first relationship
                }
            },
            {
                "key": "rel-003",
                "value": {
                    "relatedEntityId": RELATED_ID_3,
                    "startTime": "2024-01-01T00:00:00Z",  # Same timestamp as the first relationship
                    "endTime": "2024-12-31T23:59:59Z",
                    "id": "rel-003",
                    "name": "associated"  # Different type
                }
            }
        ]
    }

    # FIXME: https://github.com/LDFLK/nexoan/issues/235
    # TODO: note that the attribute temperature is a scalar value and this must be saved 
    #  as a scalar value and it should be handled as a Document type. Single key value pair. 
    #  The current implementation only supports saving tabular data.

    res = requests.post(INGESTION_API_URL, json=payload_child_1)
    assert res.status_code == 201 or res.status_code == 200, f"Failed to create entity: {res.text}"
    print("✅ Created first related entity.")

    res = requests.post(INGESTION_API_URL, json=payload_child_2)
    assert res.status_code == 201 or res.status_code == 200, f"Failed to create entity: {res.text}"
    print("✅ Created second related entity.")

    res = requests.post(INGESTION_API_URL, json=payload_child_3)
    assert res.status_code == 201 or res.status_code == 200, f"Failed to create entity: {res.text}"
    print("✅ Created third related entity.")

    res = requests.post(INGESTION_API_URL, json=payload_source)
    assert res.status_code == 201 or res.status_code == 200, f"Failed to create entity: {res.text}"
    print("✅ Created base entity for read tests.")

def test_attribute_fields_combinations():
    """Test different field combinations for attribute retrieval."""
    print("\n🔍 Testing attribute field combinations...")

    base_url = f"{READ_API_URL}/{ENTITY_ID}/attributes/employee_data"

    # Test cases with different field combinations
    test_cases = [
        {
            "name": "All fields (default)",
            "params": {"fields": []},
            "expected_fields": ["id", "e_id", "name", "age", "department", "salary"],
            "min_rows": 5
        },
        {
            "name": "ID and name only",
            "params": {"fields": ["e_id", "name"]},
            "expected_fields": ["e_id", "name"],
            "min_rows": 5
        },
        {
            "name": "Salary and department only",
            "params": {"fields": ["salary", "department"]},
            "expected_fields": ["salary", "department"],
            "min_rows": 5
        },
        {
            "name": "Single field (name)",
            "params": {"fields": ["name"]},
            "expected_fields": ["name"],
            "min_rows": 5
        },
        {
            "name": "With time range",
            "params": {
                "startTime": "2024-01-01T00:00:00Z",
                "fields": ["e_id", "name", "salary"]
            },
            "expected_fields": ["e_id", "name", "salary"],
            "min_rows": 5
        },
    ]

    for test_case in test_cases:
        print(f"  📋 Testing: {test_case['name']}")
        res = requests.get(base_url, params=test_case["params"])

        assert res.status_code == 200, f"{test_case['name']} - HTTP {res.status_code}: {res.text}"
        
        data = res.json()
        print(f"    ✅ Success: {test_case['name']}")

        # Decode and validate the response
        assert 'value' in data, f"No 'value' field in response for {test_case['name']}"
        
        value = data['value']
        decoded_data = decode_protobuf_any_value(value)

        # Validate the data - raises AssertionError if validation fails
        assert_tabular_data(decoded_data,
                          expected_columns=test_case['expected_fields'],
                          field_filter=test_case['expected_fields'],
                          min_rows=test_case['min_rows'])
        print(f"    ✅ Validation passed for {test_case['name']}")


def test_update_entity_attribute():
    """Test different field combinations for attribute retrieval."""
    print("\n🔍 Testing Update Entity Attribute...")

    attribute_name = "financial_data"

    create_payload = {
        "id": RELATED_ID_4,
        "kind": {"major": "test", "minor": "child"},
        "created": "2024-07-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-07-01T00:00:00Z",
            "endTime": "",
            "value": "Read Test Entity Child 4"
        },
        "metadata": [
            {"key": "source", "value": "unit-test-4"},
            {"key": "env", "value": "test-4"}
        ],
        "attributes": [
        ],
        "relationships": [
        ]
    }

    res = requests.post(INGESTION_API_URL, json=create_payload)
    assert res.status_code == 201 or res.status_code == 200, f"Failed to create entity: {res.text}"
    print("✅ Created first related entity.")

    update_payload = {
        "id": RELATED_ID_4,
        "attributes": [
            {
                "key": attribute_name,
                "value": {
                    "values": [
                        {
                            "startTime": "2024-08-01T00:00:00Z",
                            "endTime": "",
                            "value": {
                                "columns": ["e_id", "department", "bonus"],
                                "rows": [
                                    [1, "Engineering", 10000.50],
                                    [2, "Marketing", 65000],
                                    [3, "Sales", 15000.75],
                                    [4, "Engineering", 20000.25],
                                    [5, "Finance", 25000]
                                ]
                            }
                        }
                    ]
                }
            }
        ]
    }

    res = requests.put(f"{INGESTION_API_URL}/{RELATED_ID_4}", json=update_payload, headers={"Content-Type": "application/json"})
    assert res.status_code == 201 or res.status_code == 200, f"Failed to update entity: {res.text}"
    print("✅ Updated first related entity.")

    base_url = f"{READ_API_URL}/{RELATED_ID_4}/attributes/{attribute_name}"
    # Test cases with different field combinations
    # FIXME: NOTE THAT THE ID FIELD IS THE PRIMARY KEY THIS IS RETURNED WHEN WE ASK FOR ALL FIELDS
    #   THIS MAY NEED TO BE FIXED IN THE FUTURE. 
    test_cases = [
        {
            "name": "All fields (default)",
            "params": {"fields": []},
            "expected_fields": ["id", "e_id", "department", "bonus"],
            "min_rows": 5
        },
        {
            "name": "With time range",
            "params": {
                "startTime": "2024-01-01T00:00:00Z",
                "fields": ["department", "bonus"]
            },
            "expected_fields": ["department", "bonus"],
            "min_rows": 5
        },
    ]

    for test_case in test_cases:
        print(f"  📋 Testing: {test_case['name']}")
        res = requests.get(base_url, params=test_case["params"])

        assert res.status_code == 200, f"{test_case['name']} - HTTP {res.status_code}: {res.text}"
        
        data = res.json()
        print(f"    ✅ Success: {test_case['name']}")

        # Decode and validate the response
        assert 'value' in data, f"No 'value' field in response for {test_case['name']}"
        
        value = data['value']
        decoded_data = decode_protobuf_any_value(value)

        # Validate the data - raises AssertionError if validation fails
        assert_tabular_data(decoded_data,
                          expected_columns=test_case['expected_fields'],
                          field_filter=test_case['expected_fields'],
                          min_rows=test_case['min_rows'])
        print(f"    ✅ Validation passed for {test_case['name']}")


def test_attribute_lookup():
    """Test retrieving attributes via the read API."""
    print("\n🔍 Testing attribute retrieval...")
    
    # Test 1: Get all fields (default behavior)
    print("  📋 Testing all fields retrieval...")
    url = f"{READ_API_URL}/{ENTITY_ID}/attributes/employee_data"
    fields = []
    params = {"fields": fields}
    res = requests.get(url, params=params)
    
    assert res.status_code == 200, f"Failed to get all fields: {res.status_code} - {res.text}"
    
    data = res.json()
    print(f"    ✅ Retrieved all fields: {data}")
    
    # Decode and validate the protobuf value
    assert 'value' in data, "No 'value' field found in response"
    
    value = data['value']
    decoded_data = decode_protobuf_any_value(value)
    print(f"    ✅ Decoded data: {decoded_data}")
    
    # Validate tabular data structure
    assert_tabular_data(decoded_data, 
                      expected_columns=["id", "e_id", "name", "age", "department", "salary"],
                      min_rows=5)
    print("    ✅ All fields validation passed")
    
    # Test 2: Get specific fields only
    print("  📋 Testing specific fields retrieval...")
    fields = ["id", "name", "salary"]
    params = {"fields": fields}
    res = requests.get(url, params=params)
    
    assert res.status_code == 200, f"Failed to get specific fields: {res.status_code} - {res.text}"
    
    data = res.json()
    print(f"    ✅ Retrieved fields {fields}: {data}")
    
    # Decode the protobuf value if present
    assert 'value' in data, "No 'value' field found in response"
    
    value = data['value']
    decoded_data = decode_protobuf_any_value(value)
    print(f"    ✅ Decoded data: {decoded_data}")
    
    # Validate filtered tabular data
    assert_tabular_data(decoded_data, 
                      expected_columns=fields,
                      field_filter=fields,
                      min_rows=5)
    print("    ✅ Specific fields validation passed")
    
    # Test 3: Get fields with time range
    print("  📋 Testing fields with time range...")
    params = {
        "startTime": "2024-01-01T00:00:00Z",
        "fields": ["id", "name", "department"]
    }
    res = requests.get(url, params=params)
    
    assert res.status_code == 200, f"Failed to get filtered data: {res.status_code} - {res.text}"
    
    data = res.json()
    print(f"    ✅ Retrieved filtered data: {data}")
    
    # Decode the protobuf value if present
    assert 'value' in data, "No 'value' field found in response"
    
    value = data['value']
    decoded_data = decode_protobuf_any_value(value)
    print(f"    ✅ Decoded data: {decoded_data}")
    
    # Validate filtered tabular data
    assert_tabular_data(decoded_data, 
                      expected_columns=["id", "name", "department"],
                      field_filter=["id", "name", "department"],
                      min_rows=5)
    print("    ✅ Time-filtered fields validation passed")
    
    print("✅ Attribute lookup tests completed.")

def test_metadata_lookup():
    """Test retrieving metadata."""
    print("\n🔍 Testing metadata retrieval...")
    url = f"{READ_API_URL}/{ENTITY_ID}/metadata"
    res = requests.get(url)
    assert res.status_code == 200, f"Failed to get metadata: {res.text}"
    
    body = res.json()
    print("✅ Raw metadata response:", json.dumps(body, indent=2))
    
    # Enhanced metadata validation
    assert isinstance(body, dict), "Metadata response should be a dictionary"
    assert len(body) == 2, f"Expected 2 metadata entries, got {len(body)}"
    assert "source" in body, "Source metadata key missing"
    assert "env" in body, "Env metadata key missing"
    
    source_value = decode_protobuf_any_value(body["source"])
    env_value = decode_protobuf_any_value(body["env"])
    
    assert source_value == "unit-test", f"Source value mismatch: {source_value}"
    assert env_value == "test", f"Env value mismatch: {env_value}"

def create_government_entities():
    """Create government organizational hierarchy for search tests."""
    print("\n🟢 Creating government organizational hierarchy...")

    # Create Government entity
    gov_payload = {
        "id": GOVERNMENT_ID,
        "kind": {"major": "Organization", "minor": "Government"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Government of Sri Lanka"
        },
        "metadata": [],
        "attributes": [],
        "relationships": [
            {
                "key": "minister-rel-1",
                "value": {
                    "relatedEntityId": MINISTER_ID_1,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "",
                    "id": "gov-rel-001",
                    "name": "has_minister"
                }
            },
            {
                "key": "minister-rel-2",
                "value": {
                    "relatedEntityId": MINISTER_ID_2,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "",
                    "id": "gov-rel-002",
                    "name": "has_minister"
                }
            }
        ]
    }

    # Create Technology Minister entity
    tech_minister_payload = {
        "id": MINISTER_ID_1,
        "kind": {"major": "Organization", "minor": "Minister"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Ministry of Technology"
        },
        "metadata": [],
        "attributes": [],
        "relationships": [
            {
                "key": "dept-rel-1",
                "value": {
                    "relatedEntityId": DEPT_ID_1,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "",
                    "id": "minister-rel-001",
                    "name": "has_department"
                }
            },
            {
                "key": "dept-rel-2",
                "value": {
                    "relatedEntityId": DEPT_ID_2,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "",
                    "id": "minister-rel-002",
                    "name": "has_department"
                }
            }
        ]
    }

    # Create Health Minister entity
    health_minister_payload = {
        "id": MINISTER_ID_2,
        "kind": {"major": "Organization", "minor": "Minister"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Ministry of Health"
        },
        "metadata": [],
        "attributes": [],
        "relationships": [
            {
                "key": "dept-rel-3",
                "value": {
                    "relatedEntityId": DEPT_ID_3,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "",
                    "id": "minister-rel-003",
                    "name": "has_department"
                }
            },
            {
                "key": "dept-rel-4",
                "value": {
                    "relatedEntityId": DEPT_ID_4,
                    "startTime": "2024-01-01T00:00:00Z",
                    "endTime": "",
                    "id": "minister-rel-004",
                    "name": "has_department"
                }
            }
        ]
    }

    # Create Technology Department entities
    dept1_payload = {
        "id": DEPT_ID_1,
        "kind": {"major": "Organization", "minor": "Department"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "IT Department"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    }

    dept2_payload = {
        "id": DEPT_ID_2,
        "kind": {"major": "Organization", "minor": "Department"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Digital Services Department"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    }

    # Create Health Department entities
    dept3_payload = {
        "id": DEPT_ID_3,
        "kind": {"major": "Organization", "minor": "Department"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Hospitals Department"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    }

    dept4_payload = {
        "id": DEPT_ID_4,
        "kind": {"major": "Organization", "minor": "Department"},
        "created": "2024-01-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2024-01-01T00:00:00Z",
            "endTime": "",
            "value": "Pharmaceutical Department"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    }

    # Create all entities
    # Create departments first
    for payload in [dept1_payload, dept2_payload, dept3_payload, dept4_payload]:
        res = requests.post(INGESTION_API_URL, json=payload)
        assert res.status_code in [201, 200], f"Failed to create entity: {res.text}"
        print(f"✅ Created {payload['kind']['minor']} entity: {payload['id']}")

    # Then create ministers
    for payload in [tech_minister_payload, health_minister_payload]:
        res = requests.post(INGESTION_API_URL, json=payload)
        assert res.status_code in [201, 200], f"Failed to create entity: {res.text}"
        print(f"✅ Created {payload['kind']['minor']} entity: {payload['id']}")

    # Finally create government
    res = requests.post(INGESTION_API_URL, json=gov_payload)
    assert res.status_code in [201, 200], f"Failed to create entity: {res.text}"
    print(f"✅ Created {gov_payload['kind']['minor']} entity: {gov_payload['id']}")

def test_search_without_major_kind_or_id():
    """Test that search fails when major kind or id is not provided."""
    print("\n🔍 Testing search without major kind...")
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "minor": "Department"  # Only providing minor kind
        }
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 400, f"Search should fail without major kind: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Error response should be a dictionary"
    assert "error" in body, "Error response should contain error message"
    print("✅ Search correctly failed without major kind")

def test_search_by_kind_major():
    """Test searching entities by major kind."""
    print("\n🔍 Testing search by major kind...")
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization"
        }
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) > 0, "Expected more than 0 organizations in search response"
    # FIXME: https://github.com/LDFLK/nexoan/issues/183
    #   FIX the body number by using unique entities here or a different mechanism to make sure
    #   the test case is independent of other test cases which have been run before.
    #   Uncomment the below line to run the test case with the correct number of organizations.
    # assert len(body["body"]) > 11, "Expected 11 organizations in search response"
    
    # Verify all returned entities are of major kind "Organization"
    for entity in body["body"]:
        assert entity["kind"]["major"] == "Organization", f"Expected major kind 'Organization', got {entity['kind']['major']}"
    
    print("✅ Search by major kind successful")

def test_search_by_kind_minor():
    """Test searching entities by minor kind."""
    print("\n🔍 Testing search by minor kind...")
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization",  # Adding compulsory major kind
            "minor": "Department"
        }
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"  
    assert len(body["body"]) > 0, "Expected more than 0 departments in search response"
    # FIXME: https://github.com/LDFLK/nexoan/issues/183 
    #   FIX the body number by using unique entities here or a different mechanism to make sure
    #   the test case is independent of other test cases which have been run before.
    #   Uncomment the below line to run the test case with the correct number of departments.
    # assert len(body["body"]) > 4, "Expected 4 departments in search response"
    
    # Verify all returned entities are departments
    for entity in body["body"]:
        assert entity["kind"]["minor"] == "Department", f"Expected minor kind 'Department', got {entity['kind']['minor']}"
    
    print("✅ Search by minor kind successful")

def test_search_by_name():
    """Test searching entities by name."""
    print("\n🔍 Testing search by name...")
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization"  # Adding compulsory major kind
        },
        "name": "Ministry of Technology"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"

    print(res.text)
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 1, "Expected 1 entity in search response"
    
    # Verify the returned entity is the Technology Minister
    entity = body["body"][0]
    assert entity["id"] == MINISTER_ID_1, f"Expected minister ID {MINISTER_ID_1}, got {entity['id']}"
    assert entity["kind"]["minor"] == "Minister", f"Expected minor kind 'Minister', got {entity['kind']['minor']}"
    
    print("✅ Search by name successful")

def test_search_by_created_date():
    """Test searching entities by creation date."""
    print("\n🔍 Testing search by creation date...")
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization"  # Adding compulsory major kind
        },
        "created": "2024-01-01T00:00:00Z"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 7, "Expected 7 entities created on the same date"
    
    # Verify all returned entities have the same creation date
    for entity in body["body"]:
        assert entity["created"] == "2024-01-01T00:00:00Z", f"Expected creation date '2024-01-01T00:00:00Z', got {entity['created']}"
    
    print("✅ Search by creation date successful")

def test_search_by_name_and_kind():
    """Test searching entities by both name and kind."""
    print("\n🔍 Testing search by name and kind...")
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization",
            "minor": "Minister"
        },
        "name": "Ministry of Technology"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 1, "Expected 1 entity in search response"
    
    # Verify the returned entity matches both filters
    entity = body["body"][0]
    assert entity["id"] == MINISTER_ID_1, f"Expected minister ID {MINISTER_ID_1}, got {entity['id']}"
    assert entity["kind"]["minor"] == "Minister", f"Expected minor kind 'Minister', got {entity['kind']['minor']}"
    
    print("✅ Search by name and kind successful")

def test_search_by_kind_and_created_date():
    """Test searching entities by both kind and creation date."""
    print("\n🔍 Testing search by kind and creation date...")
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization",
            "minor": "Department"
        },
        "created": "2024-01-01T00:00:00Z"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 4, "Expected 4 departments in search response"
    
    # Verify all returned entities match both filters
    for entity in body["body"]:
        assert entity["kind"]["minor"] == "Department", f"Expected minor kind 'Department', got {entity['kind']['minor']}"
        assert entity["created"] == "2024-01-01T00:00:00Z", f"Expected creation date '2024-01-01T00:00:00Z', got {entity['created']}"
    
    print("✅ Search by kind and creation date successful")

def test_search_by_name_kind_and_created_date():
    """Test searching entities by name, kind, and creation date."""
    print("\n🔍 Testing search by name, kind, and creation date...")
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization",
            "minor": "Department"
        },
        "name": "IT Department",
        "created": "2024-01-01T00:00:00Z"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 1, "Expected 1 entity in search response"
    
    # Verify the returned entity matches all filters
    entity = body["body"][0]
    assert entity["id"] == DEPT_ID_1, f"Expected department ID {DEPT_ID_1}, got {entity['id']}"
    assert entity["kind"]["minor"] == "Department", f"Expected minor kind 'Department', got {entity['kind']['minor']}"
    assert entity["created"] == "2024-01-01T00:00:00Z", f"Expected creation date '2024-01-01T00:00:00Z', got {entity['created']}"
    
    print("✅ Search by name, kind, and creation date successful")

def test_search_by_partial_name_match():
    """Test that searching with a partial name match returns matching entities."""
    print("\n🔍 Testing search by partial name match (should return results)...")
    
    # Test entity IDs
    entity1_id = "test_partial_name_match_1"
    entity2_id = "test_partial_name_match_2"
    entity3_id = "test_partial_name_match_3"
    
    # Create test entities with different names
    entity1_payload = {
        "id": entity1_id,
        "kind": {"major": "Person", "minor": "Employee"},
        "created": "2025-04-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2025-04-01T00:00:00Z",
            "endTime": "",
            "value": "Alice Martinez"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    }
    
    entity2_payload = {
        "id": entity2_id,
        "kind": {"major": "Person", "minor": "Employee"},
        "created": "2025-04-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2025-04-01T00:00:00Z",
            "endTime": "",
            "value": "Alice Morrison"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    }
    
    entity3_payload = {
        "id": entity3_id,
        "kind": {"major": "Person", "minor": "Employee"},
        "created": "2025-04-01T00:00:00Z",
        "terminated": "",
        "name": {
            "startTime": "2025-04-01T00:00:00Z",
            "endTime": "",
            "value": "Benjamin Martinez"
        },
        "metadata": [],
        "attributes": [],
        "relationships": []
    }
    
    # Create entities
    for i, payload in enumerate([entity1_payload, entity2_payload, entity3_payload], 1):
        res = requests.post(INGESTION_API_URL, json=payload)
        assert res.status_code in [201, 200], f"Failed to create entity{i}: {res.text}"
    print(f"✅ Created test entities for partial name match test")
    
    # Search for partial name "Alice" - should match entity1 and entity2
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "major": "Person"
        },
        "name": "Alice"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    
    # Should find at least entity1 and entity2 (both contain "Alice")
    found_ids = [entity["id"] for entity in body["body"]]
    assert entity1_id in found_ids, f"Expected to find entity1 ({entity1_id})"
    assert entity2_id in found_ids, f"Expected to find entity2 ({entity2_id})"
    assert entity3_id not in found_ids, f"Entity3 ({entity3_id}) should not match 'Alice'"
    
    # Verify the names contain "Alice"
    for entity in body["body"]:
        if entity["id"] in [entity1_id, entity2_id]:
            # Decode the name from protobuf
            name_obj = json.loads(entity["name"])
            hex_value = name_obj["value"]
            decoded_name = bytes.fromhex(hex_value).decode('utf-8')
            assert "Alice" in decoded_name, f"Entity {entity['id']} name '{decoded_name}' should contain 'Alice'"
    
    print(f"✅ Found {len(body['body'])} entities matching partial name 'Alice'")
    print("✅ Search by partial name match successful")

def test_search_by_terminated_date():
    """Test searching entities by termination date."""
    print("\n🔍 Testing search by termination date...")
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization"
        },
        "terminated": "2024-12-31T23:59:59Z"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 0, "Expected 0 terminated entities in search response"
    
    print("✅ Search by termination date successful")

def test_search_by_active_entities():
    """Test searching for active (non-terminated) entities."""
    print("\n🔍 Testing search for active entities...")
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization"
        },
        "terminated": ""  # Empty terminated date means active entities
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) > 0, "Expected more than 0 active entities in search response"
    # FIXME: https://github.com/LDFLK/nexoan/issues/183 
    #   FIX the body number by using unique entities here or a different mechanism to make sure
    #   the test case is independent of other test cases which have been run before.
    #   Uncomment the below line to run the test case with the correct number of active entities.
    # assert len(body["body"]) > 7, "Expected 7 active entities in search response"
    
    # Verify all returned entities are active
    for entity in body["body"]:
        assert entity["terminated"] == "", f"Expected empty terminated date, got {entity['terminated']}"
    
    print("✅ Search for active entities successful")

def test_search_by_kind_and_terminated():
    """Test searching entities by both kind and termination status."""
    print("\n🔍 Testing search by kind and termination status...")
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization",
            "minor": "Department"
        },
        "terminated": ""  # Active departments
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) > 0, "Expected more than 0 active departments in search response"
    # FIXME: https://github.com/LDFLK/nexoan/issues/183 
    #   FIX the body number by using unique entities here or a different mechanism to make sure
    #   the test case is independent of other test cases which have been run before.
    #   Uncomment the below line to run the test case with the correct number of active departments.
    # assert len(body["body"]) > 4, "Expected 4 active departments in search response"
    
    # Verify all returned entities are active departments
    for entity in body["body"]:
        assert entity["kind"]["minor"] == "Department", f"Expected minor kind 'Department', got {entity['kind']['minor']}"
        assert entity["terminated"] == "", f"Expected empty terminated date, got {entity['terminated']}"
    
    print("✅ Search by kind and termination status successful")

def test_search_by_name_kind_and_terminated():
    """Test searching entities by name, kind, and termination status."""
    print("\n🔍 Testing search by name, kind, and termination status...")
    url = f"{READ_API_URL}/search"
    payload = {
        "kind": {
            "major": "Organization",
            "minor": "Minister"
        },
        "name": "Ministry of Technology",
        "terminated": ""  # Active minister
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 1, "Expected 1 active minister in search response"
    
    # Verify the returned entity matches all filters
    entity = body["body"][0]
    assert entity["id"] == MINISTER_ID_1, f"Expected minister ID {MINISTER_ID_1}, got {entity['id']}"
    assert entity["kind"]["minor"] == "Minister", f"Expected minor kind 'Minister', got {entity['kind']['minor']}"
    assert entity["terminated"] == "", f"Expected empty terminated date, got {entity['terminated']}"
    
    print("✅ Search by name, kind, and termination status successful")

def test_search_by_id():
    """Test searching entities by ID."""
    print("\n🔍 Testing search by ID...")
    url = f"{READ_API_URL}/search"
    payload = {
        "id": DEPT_ID_1,  # Using the IT Department ID
        # "kind": {
        #     "major": ""  # Required field
        # }
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 1, "Expected exactly one entity in search response"
    
    # Verify the returned entity matches the ID
    entity = body["body"][0]
    assert entity["id"] == DEPT_ID_1, f"Expected department ID {DEPT_ID_1}, got {entity['id']}"
    assert entity["kind"]["minor"] == "Department", f"Expected minor kind 'Department', got {entity['kind']['minor']}"
        
    # Parse the name JSON string and decode the hex value directly
    name_obj = json.loads(entity["name"])
    hex_value = name_obj["value"]
    decoded_name = bytes.fromhex(hex_value).decode('utf-8')
    
    assert decoded_name == "IT Department", f"Expected name 'IT Department', got {decoded_name}"
    
    print("✅ Search by ID successful")

def test_search_by_id_not_found():
    """Test searching for a non-existent ID."""
    print("\n🔍 Testing search by non-existent ID...")
    url = f"{READ_API_URL}/search"
    payload = {
        "id": "non-existent-id",
        # "kind": {
        #     "major": ""  # Required field
        # }
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 0, "Expected empty list for non-existent ID"
    
    print("✅ Search by non-existent ID returned empty results as expected")

def test_search_by_id_with_other_filters():
    """Test that other filters are ignored when searching by ID."""
    print("\n🔍 Testing search by ID with additional filters...")
    url = f"{READ_API_URL}/search"
    payload = {
        "id": DEPT_ID_1,
        "kind": {
            "major": "Organization", 
            "minor": "Minister"  # This should be ignored since we're searching by ID
        },
        "name": "Wrong Name",  # This should be ignored
        "created": "2023-01-01T00:00:00Z"  # This should be ignored
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Search failed: {res.text}"
    
    body = res.json()
    assert isinstance(body, dict), "Search response should be a dictionary"
    assert "body" in body, "Search response should have a 'body' field"
    assert isinstance(body["body"], list), "Search response body should be a list"
    assert len(body["body"]) == 1, "Expected exactly one entity in search response"
    
    # Verify the returned entity matches the ID despite other filters
    entity = body["body"][0]
    assert entity["id"] == DEPT_ID_1, f"Expected department ID {DEPT_ID_1}, got {entity['id']}"
    assert entity["kind"]["minor"] == "Department", f"Expected minor kind 'Department', got {entity['kind']['minor']}"
    
    # Decode the name value from protobuf Any type
    name_obj = json.loads(entity["name"])
    hex_value = name_obj["value"]
    decoded_name = bytes.fromhex(hex_value).decode('utf-8')
    assert decoded_name == "IT Department", f"Expected name 'IT Department', got {decoded_name}"
    
    print("✅ Search by ID ignored additional filters as expected")

def test_relations_no_filters():
    """Test /relations endpoint with no filters (should return all relationships for the entity)."""
    print("\n🔍 Testing /relations with no filters...")
    url = f"{READ_API_URL}/{ENTITY_ID}/relations"
    res = requests.post(url, json={})
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert isinstance(body, list), "Response should be a list"
    assert len(body) >= 3, "Expected at least 3 relationships"
    print("✅ /relations with no filters:", json.dumps(body, indent=2))

def test_relations_filter_by_name():
    """Test /relations endpoint filtering by relationship name."""
    print("\n🔍 Testing /relations with filter by name...")
    url = f"{READ_API_URL}/{ENTITY_ID}/relations"
    payload = {"name": "linked"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["name"] == "linked" for rel in body), "All relationships should have name 'linked'"
    print("✅ /relations with name filter:", json.dumps(body, indent=2))

def test_relations_filter_by_related_entity_id():
    """Test /relations endpoint filtering by relatedEntityId."""
    print("\n🔍 Testing /relations with filter by relatedEntityId...")
    url = f"{READ_API_URL}/{ENTITY_ID}/relations"
    payload = {"relatedEntityId": RELATED_ID_1}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["relatedEntityId"] == RELATED_ID_1 for rel in body), "All relationships should have the correct relatedEntityId"
    print("✅ /relations with relatedEntityId filter:", json.dumps(body, indent=2))

def test_relations_filter_by_start_time():
    """Test /relations endpoint filtering by startTime."""
    print("\n🔍 Testing /relations with filter by startTime...")
    url = f"{READ_API_URL}/{ENTITY_ID}/relations"
    payload = {"startTime": "2024-06-01T00:00:00Z"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["startTime"] == "2024-06-01T00:00:00Z" for rel in body), "All relationships should have the correct startTime"
    print("✅ /relations with startTime filter:", json.dumps(body, indent=2))

def test_relations_filter_by_end_time():
    """Test /relations endpoint filtering by endTime."""
    print("\n🔍 Testing /relations with filter by endTime...")
    url = f"{READ_API_URL}/{ENTITY_ID}/relations"
    payload = {"endTime": "2024-12-31T23:59:59Z"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["endTime"] == "2024-12-31T23:59:59Z" for rel in body), "All relationships should have the correct endTime"
    print("✅ /relations with endTime filter:", json.dumps(body, indent=2))

def test_relations_filter_by_multiple_fields():
    """Test /relations endpoint filtering by multiple fields."""
    print("\n🔍 Testing /relations with multiple filters...")
    url = f"{READ_API_URL}/{ENTITY_ID}/relations"
    payload = {
        "name": "linked",
        "relatedEntityId": RELATED_ID_2,
        "startTime": "2024-06-01T00:00:00Z"
    }
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert len(body) == 1, "Expected exactly one relationship"
    rel = body[0]
    assert rel["name"] == "linked"
    assert rel["relatedEntityId"] == RELATED_ID_2
    assert rel["startTime"] == "2024-06-01T00:00:00Z"
    print("✅ /relations with multiple filters:", json.dumps(body, indent=2))

def test_relations_filter_nonexistent():
    """Test /relations endpoint with filters that match nothing."""
    print("\n🔍 Testing /relations with non-existent filter...")
    url = f"{READ_API_URL}/{ENTITY_ID}/relations"
    payload = {"name": "nonexistent"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert isinstance(body, list), "Response should be a list"
    assert len(body) == 0, "Expected no relationships for non-existent filter"
    print("✅ /relations with non-existent filter returned empty list.")

def test_relations_filter_by_active_at():
    """Test /relations endpoint filtering by activeAt only."""
    print("\n🔍 Testing /relations with filter by activeAt...")
    url = f"{READ_API_URL}/{ENTITY_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    # Should return relationships active at this time
    assert isinstance(body, list), "Response should be a list"
    assert any(rel["id"] == "rel-001" for rel in body), "Expected rel-001 to be active at 2024-07-01T00:00:00Z"
    assert any(rel["id"] == "rel-002" for rel in body), "Expected rel-002 to be active at 2024-07-01T00:00:00Z"
    assert any(rel["id"] == "rel-003" for rel in body), "Expected rel-003 to be active at 2024-07-01T00:00:00Z"
    print("✅ /relations with activeAt filter:", json.dumps(body, indent=2))


def test_relations_filter_by_active_at_and_name():
    """Test /relations endpoint filtering by activeAt and name."""
    print("\n🔍 Testing /relations with filter by activeAt and name...")
    url = f"{READ_API_URL}/{ENTITY_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "name": "linked"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["name"] == "linked" for rel in body), "All relationships should have name 'linked'"
    assert any(rel["id"] == "rel-001" for rel in body), "Expected rel-001 to be present"
    assert any(rel["id"] == "rel-002" for rel in body), "Expected rel-002 to be present"
    print("✅ /relations with activeAt and name filter:", json.dumps(body, indent=2))


def test_relations_filter_by_active_at_and_related_entity_id():
    """Test /relations endpoint filtering by activeAt and relatedEntityId."""
    print("\n🔍 Testing /relations with filter by activeAt and relatedEntityId...")
    url = f"{READ_API_URL}/{ENTITY_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "relatedEntityId": RELATED_ID_1}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["relatedEntityId"] == RELATED_ID_1 for rel in body), "All relationships should have the correct relatedEntityId"
    assert any(rel["id"] == "rel-001" for rel in body), "Expected rel-001 to be present"
    print("✅ /relations with activeAt and relatedEntityId filter:", json.dumps(body, indent=2))


def test_relations_filter_by_active_at_and_direction():
    """Test /relations endpoint filtering by activeAt and direction."""
    print("\n🔍 Testing /relations with filter by activeAt and direction...")
    url = f"{READ_API_URL}/{ENTITY_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "direction": "OUTGOING"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["direction"] == "OUTGOING" for rel in body), "All relationships should have direction 'OUTGOING'"
    print("✅ /relations with activeAt and direction filter:", json.dumps(body, indent=2))


def test_relations_filter_by_active_at_and_name_and_direction():
    """Test /relations endpoint filtering by activeAt, name, and direction."""
    print("\n🔍 Testing /relations with filter by activeAt, name, and direction...")
    url = f"{READ_API_URL}/{ENTITY_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "name": "linked", "direction": "OUTGOING"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    assert all(rel["name"] == "linked" for rel in body), "All relationships should have name 'linked'"
    assert all(rel["direction"] == "OUTGOING" for rel in body), "All relationships should have direction 'OUTGOING'"
    print("✅ /relations with activeAt, name, and direction filter:", json.dumps(body, indent=2))


def test_relations_filter_by_active_at_and_time_range_invalid():
    """Test /relations endpoint with activeAt and startTime (should return 400)."""
    print("\n🔍 Testing /relations with activeAt and startTime (should fail)...")
    url = f"{READ_API_URL}/{ENTITY_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "startTime": "2024-01-01T00:00:00Z"}
    res = requests.post(url, json=payload)
    assert res.status_code == 400, f"Expected 400 Bad Request, got {res.status_code}: {res.text}"
    body = res.json()
    assert "error" in body, "Error message should be present in 400 response"
    print("✅ /relations with activeAt and startTime correctly failed:", json.dumps(body, indent=2))

def test_gov_relations_filter_by_active_at_and_direction():
    """Test /relations endpoint for government entity with activeAt and direction OUTGOING."""
    print("\n🔍 Testing /relations for government entity with activeAt and direction OUTGOING...")
    url = f"{READ_API_URL}/{GOVERNMENT_ID}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "direction": "OUTGOING"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    # Should return has_minister relationships
    assert all(rel["direction"] == "OUTGOING" for rel in body), "All relationships should be OUTGOING"
    assert all(rel["name"] == "has_minister" for rel in body), "All relationships should be has_minister"
    assert any(rel["relatedEntityId"] == MINISTER_ID_1 for rel in body), "Should include MINISTER_ID_1"
    assert any(rel["relatedEntityId"] == MINISTER_ID_2 for rel in body), "Should include MINISTER_ID_2"
    print("✅ Government OUTGOING has_minister relationships:", json.dumps(body, indent=2))

def test_minister_relations_filter_by_active_at_and_direction():
    """Test /relations endpoint for minister entity with activeAt and direction OUTGOING."""
    print("\n🔍 Testing /relations for minister entity with activeAt and direction OUTGOING...")
    url = f"{READ_API_URL}/{MINISTER_ID_1}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "direction": "OUTGOING"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    # Should return has_department relationships
    print("✅ Minister OUTGOING has_department relationships:", json.dumps(body, indent=2))
    assert all(rel["direction"] == "OUTGOING" for rel in body), "All relationships should be OUTGOING"
    assert all(rel["name"] == "has_department" for rel in body), "All relationships should be has_department"
    assert any(rel["relatedEntityId"] == DEPT_ID_1 for rel in body), "Should include DEPT_ID_1"
    assert any(rel["relatedEntityId"] == DEPT_ID_2 for rel in body), "Should include DEPT_ID_2"
    print("✅ Minister OUTGOING has_department relationships:", json.dumps(body, indent=2))

    print("\n🔍 Testing /relations for minister entity with activeAt and direction INCOMING...")
    payload = {"activeAt": "2024-07-01T00:00:00Z", "direction": "INCOMING"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    # Should return has_minister relationship from government
    assert all(rel["direction"] == "INCOMING" for rel in body), "All relationships should be INCOMING"
    assert all(rel["name"] == "has_minister" for rel in body), "All relationships should be has_minister"
    assert any(rel["relatedEntityId"] == GOVERNMENT_ID for rel in body), "Should include GOVERNMENT_ID as relatedEntityId"
    print("✅ Minister INCOMING has_minister relationship:", json.dumps(body, indent=2))


def test_department_relations_filter_by_active_at_and_direction():
    """Test /relations endpoint for department entity with activeAt and direction INCOMING."""
    print("\n🔍 Testing /relations for department entity with activeAt and direction INCOMING...")
    url = f"{READ_API_URL}/{DEPT_ID_1}/relations"
    payload = {"activeAt": "2024-07-01T00:00:00Z", "direction": "INCOMING"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    # Should return has_department relationship from minister
    assert all(rel["direction"] == "INCOMING" for rel in body), "All relationships should be INCOMING"
    assert all(rel["name"] == "has_department" for rel in body), "All relationships should be has_department"
    assert any(rel["relatedEntityId"] == MINISTER_ID_1 for rel in body), "Should include MINISTER_ID_1 as relatedEntityId"
    print("✅ Department INCOMING has_department relationship:", json.dumps(body, indent=2))

def test_minister_relations_filter_by_active_at_only():
    """Test /relations endpoint for minister entity with only activeAt (2024-01-02)."""
    print("\n🔍 Testing /relations for minister entity with only activeAt...")
    url = f"{READ_API_URL}/{MINISTER_ID_1}/relations"
    payload = {"activeAt": "2024-01-02T00:00:00Z"}
    res = requests.post(url, json=payload)
    assert res.status_code == 200, f"Failed to get relationships: {res.text}"
    body = res.json()
    # Should return both outgoing has_department and incoming has_minister relationships
    directions = set(rel["direction"] for rel in body)
    names = set(rel["name"] for rel in body)
    related_ids = set(rel["relatedEntityId"] for rel in body)
    print("✅ Minister relationships with only activeAt:", json.dumps(body, indent=2))
    assert "OUTGOING" in directions, "Should include OUTGOING relationships"
    assert "INCOMING" in directions, "Should include INCOMING relationships"
    assert "has_department" in names, "Should include has_department relationships"
    assert "has_minister" in names, "Should include has_minister relationships"
    assert MINISTER_ID_1 in related_ids or DEPT_ID_1 in related_ids or DEPT_ID_2 in related_ids or GOVERNMENT_ID in related_ids, "Should include expected related entity IDs"

if __name__ == "__main__":
    print("🚀 Running Read API E2E Tests...")

    try:
        print("Basic Read Tests...")
        print("Testing comprehensive validation...")
        test_comprehensive_validation()
        print("Testing protobuf decoding...")
        test_protobuf_decoding()
        print("Creating entity for read tests...")
        create_entity_for_read()
        print("Testing generic validation examples...")
        test_generic_validation_examples()
        print("Testing attribute field combinations...")
        test_attribute_fields_combinations()
        print("Testing attribute lookup...")
        test_attribute_lookup()
        print("Testing update entity attribute...")
        test_update_entity_attribute()
        print("Testing metadata lookup...")
        test_metadata_lookup()
        
        # Run government organization search tests
        create_government_entities()
        test_search_without_major_kind_or_id()
        test_search_by_kind_major()
        test_search_by_kind_minor()
        test_search_by_name()
        test_search_by_created_date()
        
        # Run ID-based search tests
        test_search_by_id()
        test_search_by_id_not_found()
        test_search_by_id_with_other_filters()
        
        # Run combined filter tests
        test_search_by_name_and_kind()
        test_search_by_kind_and_created_date()
        test_search_by_name_kind_and_created_date()
        test_search_by_partial_name_match()
        
        # Run terminated date filter tests
        test_search_by_terminated_date()
        test_search_by_active_entities()
        test_search_by_kind_and_terminated()
        test_search_by_name_kind_and_terminated()
        
        # Run relationship filter tests
        test_relations_no_filters()
        test_relations_filter_by_name()
        test_relations_filter_by_related_entity_id()
        test_relations_filter_by_start_time()
        test_relations_filter_by_end_time()
        test_relations_filter_by_multiple_fields()
        test_relations_filter_nonexistent()
        test_relations_filter_by_active_at()
        test_relations_filter_by_active_at_and_name()
        test_relations_filter_by_active_at_and_related_entity_id()
        test_relations_filter_by_active_at_and_direction()
        test_relations_filter_by_active_at_and_name_and_direction()
        test_relations_filter_by_active_at_and_time_range_invalid()
        test_gov_relations_filter_by_active_at_and_direction()
        test_minister_relations_filter_by_active_at_and_direction()
        test_department_relations_filter_by_active_at_and_direction()
        test_minister_relations_filter_by_active_at_only()
        
        print("\n🎉 All Read API tests passed!")
    except AssertionError as e:
        print(f"\n❌ Test failed: {e}")
        sys.exit(1)
