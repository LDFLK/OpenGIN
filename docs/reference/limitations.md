# Limitations

## Read API

1. Read for data like tables or documents (metadata, unstructured documents) doesn't include filters for querying parameters inside tables. 
2. Join, aggregations and advanced data processing queries are not yet supported. 
3. Sub-graph insertion as an attribute is not yet supported. 
4. Scalar value insertion as an attribute is not yet supported.

## OpenAPI Contract and Ballerina Service generation

1. Time is being infered as `string` field. Must implement a time data type in the protobuf specification. 