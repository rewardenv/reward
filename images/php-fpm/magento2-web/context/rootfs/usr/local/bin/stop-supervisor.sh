#!/bin/sh
printf "READY\n";

while true; do
  read -r line
  echo "Processing Event: $line" >&2;
  kill -3 1
done < /dev/stdin
