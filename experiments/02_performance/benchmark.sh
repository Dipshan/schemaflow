#!/bin/bash
echo "Performance Benchmark — $(date)" > results.txt
echo "=================================" >> results.txt

QUERY='{"query": "SELECT tconst, primary_title FROM title_basics WHERE start_year > 2000"}'

echo "Running 1000 iterations of /rewrite..." >> results.txt

for i in $(seq 1 10000); do
    START=$(date +%s%N)
    curl -s -X POST http://localhost:8080/rewrite \
        -H "Content-Type: application/json" \
        -d "$QUERY" >> /dev/null
    END=$(date +%s%N)
    DURATION=$(( (END - START) / 1000000 ))
    echo "Iteration $i: ${DURATION}ms" >> results.txt
done

echo "" >> results.txt
echo "Benchmark complete." >> results.txt