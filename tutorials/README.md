# OpenGIN Tutorials

This directory contains example applications and data to help you get started with OpenGIN.

## Personal Life Ecology Example

The `my_first_opengin_app.py` script demonstrates how to ingest a rich dataset representing a "Personal Life Ecology" into OpenGIN. It models a person ("Alex"), their relationships (Work, Gym, Home), and their financial life (Income, Expenses).

### Prerequisites

1.  **Python 3.8+**
2.  **OpenGIN Running Locally**:
    - Ingestion API at `http://localhost:8080`
    - Read API at `http://localhost:8081`
3.  **Python Dependencies**:
    ```bash
    pip install pandas requests
    ```

### Directory Structure

```
tutorials/
├── my_first_opengin_app.py  # The main ingestion script
├── data/                    # CSV and JSON data files
│   ├── entities.csv         # Core entities (Alex, TechCorp, etc.)
│   ├── relationships.csv    # Links between entities
│   ├── expenses.csv         # Tabular data linked to entities
│   ├── income.csv           # Tabular data linked to entities
│   └── metadata.json        # Extra metadata for entities
└── README.md                # This file
```

### Running the Example

1.  Navigate to the project root or the `tutorials` directory.
3.  Run the script using the CLI commands:

    **Create / Ingest Data:**
    ```bash
    python tutorials/my_first_opengin_app.py create
    ```

    **Verify Ingestion:**
    ```bash
    python tutorials/my_first_opengin_app.py verify
    ```

    *Note: If running from inside the `tutorials` directory, adjust the path (e.g., `python my_first_opengin_app.py create`).*

### Expected Output

The script will:
1.  Load data from the `data/` directory.
2.  Ingest entities and relationships into OpenGIN.
3.  Print a success message for each entity.
4.  **Verify** the ingestion by querying the OpenGIN Read API and finding "Person" entities.

```text
📂 Loading data files...
✅ Loaded 7 entities, 6 relationships

🚀 Starting Ingestion...
  Processing Entity: person-alex (Alex)...
    ✅ Success
  ...
  
🔍 Verifying Ingestion...
✅ Found 1 Person entities.
   - person-alex (Individual)
```
