# A/B Live Test Report

- Generated: 2026-04-18T09:40:54Z

| Name | URL | Before | After |
|---|---|---|---|
| amazon-home | https://www.amazon.com | status=424 blocked=true attempts=1 ms=3174 cost=0.0001 err="" | status=424 blocked=false attempts=1 ms=7906 cost=0.0048 err="" |
| amazon-search-page2 | https://www.amazon.com/s?k=laptop&page=2 | status=503 blocked=true attempts=1 ms=1591 cost=0.0001 err="" | status=424 blocked=false attempts=2 ms=6368 cost=0.0064 err="" |
| walmart-home | https://www.walmart.com | status=200 blocked=true attempts=1 ms=19487 cost=0.0016 err="" | status=200 blocked=true attempts=3 ms=44774 cost=0.0138 err="request remained blocked after challenge orchestration retries" |
| walmart-search-page2 | https://www.walmart.com/search?q=laptop&page=2 | status=200 blocked=true attempts=1 ms=8651 cost=0.0004 err="" | status=200 blocked=true attempts=3 ms=22618 cost=0.0027 err="request remained blocked after challenge orchestration retries" |
| target-home | https://www.target.com | status=200 blocked=true attempts=1 ms=6815 cost=0.0014 err="" | status=200 blocked=true attempts=3 ms=18994 cost=0.0105 err="request remained blocked after challenge orchestration retries" |
| target-search-page2 | https://www.target.com/s?searchTerm=laptop&page=2 | status=200 blocked=true attempts=1 ms=11487 cost=0.0023 err="" | status=200 blocked=true attempts=3 ms=22215 cost=0.0115 err="request remained blocked after challenge orchestration retries" |

## After Debug Trace

### amazon-home

- `attempt=1 proxy={us } blocked=false err=false health=map[{ca }:0 {us }:0] delta=map[{ca }:0 {us }:0]`

### amazon-search-page2

- `attempt=1 proxy={ca } blocked=false err=true health=map[{ca }:0 {us }:0] delta=map[{ca }:0 {us }:0]`
- `attempt=2 proxy={us } blocked=false err=false health=map[{ca }:2 {us }:0] delta=map[{ca }:2 {us }:0]`

### walmart-home

- `attempt=1 proxy={us } blocked=true err=false health=map[{ca }:0 {us }:0] delta=map[{ca }:0 {us }:0]`
- `attempt=2 proxy={ca } blocked=false err=true health=map[{ca }:0 {us }:1] delta=map[{ca }:0 {us }:1]`
- `attempt=3 proxy={us } blocked=true err=false health=map[{ca }:2 {us }:1] delta=map[{ca }:2 {us }:0]`

### walmart-search-page2

- `attempt=1 proxy={ca } blocked=false err=true health=map[{ca }:2 {us }:2] delta=map[{ca }:2 {us }:2]`
- `attempt=2 proxy={us } blocked=true err=false health=map[{ca }:4 {us }:2] delta=map[{ca }:2 {us }:0]`
- `attempt=3 proxy={us } blocked=true err=false health=map[{ca }:4 {us }:3] delta=map[{ca }:0 {us }:1]`

### target-home

- `attempt=1 proxy={us } blocked=true err=false health=map[{ca }:0 {us }:0] delta=map[{ca }:0 {us }:0]`
- `attempt=2 proxy={ca } blocked=false err=true health=map[{ca }:0 {us }:1] delta=map[{ca }:0 {us }:1]`
- `attempt=3 proxy={us } blocked=true err=false health=map[{ca }:2 {us }:1] delta=map[{ca }:2 {us }:0]`

### target-search-page2

- `attempt=1 proxy={ca } blocked=false err=true health=map[{ca }:2 {us }:2] delta=map[{ca }:2 {us }:2]`
- `attempt=2 proxy={us } blocked=true err=false health=map[{ca }:4 {us }:2] delta=map[{ca }:2 {us }:0]`
- `attempt=3 proxy={us } blocked=true err=false health=map[{ca }:4 {us }:3] delta=map[{ca }:0 {us }:1]`

