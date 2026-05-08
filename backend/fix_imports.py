#!/usr/bin/env python3
import subprocess
import sys
import os

def fix_imports(file_path):
    with open(file_path, 'r') as f:
        content = f.read()

    content = content.replace(
        '"github.com/portfolio-sim/backend/internal/logging"',
        '"github.com/portfolio-sim/backend/internal/logging"'
    ).replace(
        '"github.com/portfolio-sim/backend/internal/models"',
        '"github.com/portfolio-sim/backend/internal/models"'
    )

    with open(file_path, 'w') as f:
        f.write(content)

    print(f"Fixed imports in {file_path}")

def main():
    for root, dirs, files in os.walk('backend'):
        for file in files:
            if file.endswith('.go'):
                path = os.path.join(root, file)
                fix_imports(path)

if __name__ == '__main__':
    main()
