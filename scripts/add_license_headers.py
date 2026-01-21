# Copyright 2025 Lanka Data Foundation
# SPDX-License-Identifier: Apache-2.0

import os
import argparse

# License headers
LICENSE_HEADER_SLASH = """// Copyright 2025 Lanka Data Foundation
// SPDX-License-Identifier: Apache-2.0

"""

LICENSE_HEADER_HASH = """# Copyright 2025 Lanka Data Foundation
# SPDX-License-Identifier: Apache-2.0

"""

# Extensions mapping
EXTENSIONS = {
    '.go': LICENSE_HEADER_SLASH,
    '.bal': LICENSE_HEADER_SLASH,
    '.py': LICENSE_HEADER_HASH,
    '.sh': LICENSE_HEADER_HASH,
    '.proto': LICENSE_HEADER_SLASH,
}

# Special filenames
SPECIAL_FILES = {
    'Dockerfile': LICENSE_HEADER_HASH
}

def has_license(content):
    """Check if the file already has a license header."""
    # Check within the first 10 lines to avoid reading the whole file if it's large.
    header_check_text = "\n".join(content.splitlines()[:10])
    return "Copyright" in header_check_text and "SPDX-License-Identifier" in header_check_text

def get_header(filename):
    """Get the appropriate header for a given filename."""
    _, ext = os.path.splitext(filename)
    if filename in SPECIAL_FILES:
        return SPECIAL_FILES[filename]
    return EXTENSIONS.get(ext)

def process_file(filepath, dry_run=False):
    """Process a single file to add the license header."""
    filename = os.path.basename(filepath)
    header = get_header(filename)
    
    if not header:
        return

    try:
        with open(filepath, 'r', encoding='utf-8') as f:
            content = f.read()

        if has_license(content):
            print(f"[SKIP] Existing license: {filepath}")
            return

        if dry_run:
            print(f"[DRY-RUN] Would add license to: {filepath}")
            return

        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(header + content)
        print(f"[b] Added license to: {filepath}")

    except Exception as e:
        print(f"[ERROR] Failed to process {filepath}: {e}")

def process_directory(directory, dry_run=False):
    """Recursively process a directory."""
    for root, dirs, files in os.walk(directory):
        for file in files:
            filepath = os.path.join(root, file)
            process_file(filepath, dry_run)

def main():
    parser = argparse.ArgumentParser(description="Add Apache 2.0 license headers to source files.")
    parser.add_argument("directories", nargs='+', help="Directories to process")
    parser.add_argument("--dry-run", action="store_true", help="Print actions without modifying files")
    
    args = parser.parse_args()

    for directory in args.directories:
        if os.path.exists(directory):
            print(f"Processing directory: {directory}")
            process_directory(directory, args.dry_run)
        else:
            print(f"Directory not found: {directory}")

if __name__ == "__main__":
    main()
