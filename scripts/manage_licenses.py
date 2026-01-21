# Copyright 2025 Lanka Data Foundation
# SPDX-License-Identifier: Apache-2.0

import argparse
import sys
import os

# Ensure the script can import add_license_headers from the same directory
sys.path.append(os.path.dirname(os.path.abspath(__file__)))
from add_license_headers import process_directory

COMPONENT_MAP = {
    "core": "opengin/core-api",
    "ingestion": "opengin/ingestion-api",
    "read": "opengin/read-api",
    "perf": "perf",
    "tutorials": "tutorials",
    "deployment": "deployment",
}

def main():
    parser = argparse.ArgumentParser(
        description="Manage Apache 2.0 license headers for OpenGIN components."
    )
    
    parser.add_argument(
        "components",
        nargs="*",
        help="List of components to process (e.g., 'core', 'ingestion'). Use 'all' for everything.",
    )
    
    parser.add_argument(
        "--dry-run",
        action="store_true",
        help="Preview changes without modifying files.",
    )
    
    parser.add_argument(
        "--list",
        action="store_true",
        help="List all available components and their paths.",
    )

    args = parser.parse_args()

    # Handle --list
    if args.list:
        print("Available Components:")
        for name, path in COMPONENT_MAP.items():
            print(f"  {name:<12} -> {path}")
        return

    # Handle no components specified
    if not args.components:
        parser.print_help()
        sys.exit(1)

    # Determine components to process
    targets = []
    if "all" in args.components:
        targets = list(COMPONENT_MAP.keys())
    else:
        for name in args.components:
            if name in COMPONENT_MAP:
                targets.append(name)
            else:
                print(f"[ERROR] Invalid component: '{name}'. Use --list to see available options.")
                sys.exit(1)

    # Process targets
    base_dir = os.getcwd() # Assuming run from repo root
    # Adjust base_dir check if script is run from scripts/ dir
    # Ideally script is run from repo root as `python scripts/manage_licenses.py`
    
    print(f"Processing components: {', '.join(targets)}")
    if args.dry_run:
        print("[NOTICE] Running in DRY-RUN mode. No files will be modified.")

    for name in targets:
        rel_path = COMPONENT_MAP[name]
        full_path = os.path.join(base_dir, rel_path)
        
        if os.path.exists(full_path):
            print(f"\n--- Component: {name} ({rel_path}) ---")
            process_directory(full_path, dry_run=args.dry_run)
        else:
            print(f"\n[WARNING] Component path not found: {rel_path} (Skipping)")

if __name__ == "__main__":
    main()
