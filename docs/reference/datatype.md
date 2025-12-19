# Type Inference System

The type inference system is a core component of the LDF Architecture, responsible for automatically determining the appropriate data types for values in the data structure. This document explains how the system works, what types are supported, and how type inference rules are applied.

## Supported Data Types

### Primitive Types

1. **Integer (`int`)**
   - Whole numbers without decimal points
   - Examples: `42`, `-1`, `0`
   - Used for counting, indexing, and whole number quantities

2. **Float (`float`)**
   - Numbers with decimal points or scientific notation
   - Examples: `3.14`, `-0.001`, `1.0e-10`
   - Used for measurements, percentages, and precise calculations

3. **String (`string`)**
   - Text data of any length
   - Examples: `"hello"`, `"user123"`, `""`
   - Used for names, descriptions, and general text

4. **Boolean (`bool`)**
   - True/false values
   - Examples: `true`, `false`
   - Used for flags, conditions, and binary states

5. **Null (`null`)**
   - Represents absence of a value
   - Example: `null`
   - Used when a value is not provided or not applicable

### Special Types

1. **Date (`date`)**
   - Calendar dates without time information
   - Supported formats:
     - `YYYY-MM-DD` (e.g., "2024-03-20")
     - `DD/MM/YYYY` (e.g., "20/03/2024")
     - `MM/DD/YYYY` (e.g., "03/20/2024")
     - `YYYY.MM.DD` (e.g., "2024.03.20")
     - `DD-MM-YYYY` (e.g., "20-03-2024")
     - `MM-DD-YYYY` (e.g., "03-20-2024")
     - `YYYY/MM/DD` (e.g., "2024/03/20")

2. **Time (`time`)**
   - Time of day without date information
   - Supported formats:
     - `HH:MM:SS` (e.g., "14:30:00")
     - `HH:MM` (e.g., "14:30")
     - `h:MM AM/PM` (e.g., "2:30 PM")
     - `HH:MM:SS.mmm` (e.g., "14:30:00.000")
     - `HH:MM:SS±HH:MM` (e.g., "14:30:00-07:00")
     - `HH:MM:SSZ` (e.g., "14:30:00Z")

3. **DateTime (`datetime`)**
   - Combined date and time information
   - Supported formats:
     - RFC3339 (e.g., "2024-03-20T14:30:00Z07:00")
     - `YYYY-MM-DD HH:MM:SS` (e.g., "2024-03-20 14:30:00")
     - `YYYY-MM-DDTHH:MM:SS` (e.g., "2024-03-20T14:30:00")
     - `DD/MM/YYYY HH:MM:SS` (e.g., "20/03/2024 14:30:00")
     - `MM/DD/YYYY HH:MM:SS` (e.g., "03/20/2024 14:30:00")
     - `YYYY.MM.DD HH:MM:SS` (e.g., "2024.03.20 14:30:00")
     - `YYYY-MM-DD HH:MM:SS.mmm` (e.g., "2024-03-20 14:30:00.000")

## Type Inference Rules

The system follows these rules when inferring types:

1. **Number Type Resolution**
   - If a number has a decimal point or is in scientific notation, it's inferred as `float`
   - If a number is a whole number (no decimal), it's inferred as `int`
   - Special case: zero (0) is checked against its string representation to determine if it was originally a float

2. **String Type Resolution**
   - All text values are first checked against date/time patterns
   - If the text matches a date pattern, it's inferred as `date`
   - If the text matches a time pattern, it's inferred as `time`
   - If the text matches a datetime pattern, it's inferred as `datetime`
   - Otherwise, it's inferred as `string`

3. **Null Handling**
   - Explicit null values are inferred as `null` type
   - The `null` type is always nullable
   - Missing fields are treated as `null`

4. **Array Type Resolution**
   - Arrays are marked with `IsArray: true`
   - The element type is determined from the first non-null element
   - Empty arrays default to `string` element type
   - Array type information includes:
     ```json
     {
       "type": "string",
       "is_array": true,
       "array_type": {
         "type": "int"
       }
     }
     ```

## Examples

### Basic Types
```json
{
  "integer_value": 42,
  "float_value": 3.14,
  "string_value": "hello",
  "boolean_value": true,
  "null_value": null
}
```

### Special Types
```json
{
  "date_value": "2024-03-20",
  "time_value": "14:30:00",
  "datetime_value": "2024-03-20T14:30:00Z"
}
```

### Array Types
```json
{
  "int_array": [1, 2, 3],
  "mixed_array": ["a", 1, true],
  "empty_array": []
}
```

## Type Information Structure

The type information is structured to define the fundamental data type, including flags for nullability and array structures. For nested types, it recursively specifies element types for arrays and property maps for objects, enabling comprehensive type definitions for complex data structures.

## Best Practices

1. **Date and Time Formatting**
   - Use ISO 8601 / RFC 3339 formats when possible
   - Include timezone information for datetime values
   - Be consistent with format choice within your application

2. **Number Handling**
   - Use integers for counting and indexing
   - Use floats for measurements and calculations
   - Be explicit about decimal points when floating-point precision is required

3. **Null Values**
   - Use explicit null values rather than empty strings or zero values
   - Document which fields are nullable in your schema
   - Consider using nullable types in strongly-typed languages

4. **Array Types**
   - Keep array elements consistent in type
   - Provide type information for empty arrays
   - Consider using single-type arrays for better type safety
