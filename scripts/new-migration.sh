#!/bin/sh

timestamp=$(date +"%Y%m%d%H%M%S")
suffix=$1

# Construct the file name
if [ -n "$suffix" ]; then
    filename="db/migrations/${timestamp}_${suffix}.sql"
else
    filename="db/migrations/${timestamp}.sql"
fi

touch "$filename"