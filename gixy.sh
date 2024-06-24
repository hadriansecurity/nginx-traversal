#!/bin/bash

# Find all .conf files and iterate over them
find . -name '*.conf' | while IFS= read -r file; do
    echo "$file"  # Echo the file name
    gixy -f json "$file"  # Execute gixy on the file
done
