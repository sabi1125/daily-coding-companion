#!/bin/sh

echo "Piston core installed. Installing languages..."

RUNTIMES="
python=3.12.0
node=20.11.1
go=1.16.2
gcc=10.2.0
"

for pair in $RUNTIMES; do
  lang=$(echo "$pair" | cut -d= -f1)
  version=$(echo "$pair" | cut -d= -f2)

  echo "Installing lang: $lang ($version)"
  curl -s -X POST http://piston:2000/api/v2/packages \
    -H "Content-Type: application/json" \
    -d "{\"language\": \"$lang\", \"version\": \"$version\"}"
done

echo "All languages installed"
