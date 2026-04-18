# A/B Live Test Report

- Generated: 2026-04-18T09:36:34Z

| Name | URL | Before | After |
|---|---|---|---|
| amazon-home | https://www.amazon.com | status=200 blocked=true attempts=1 ms=3089 cost=0.0001 err="" | status=424 blocked=false attempts=1 ms=7764 cost=0.0043 err="" |
| amazon-search-page2 | https://www.amazon.com/s?k=laptop&page=2 | status=503 blocked=true attempts=1 ms=1746 cost=0.0001 err="" | status=424 blocked=false attempts=2 ms=6675 cost=0.0066 err="" |
| walmart-home | https://www.walmart.com | status=200 blocked=true attempts=1 ms=19350 cost=0.0016 err="" | status=200 blocked=true attempts=3 ms=44743 cost=0.0138 err="request remained blocked after challenge orchestration retries" |
| walmart-search-page2 | https://www.walmart.com/search?q=laptop&page=2 | status=200 blocked=true attempts=1 ms=11049 cost=0.0005 err="" | status=200 blocked=true attempts=3 ms=24061 cost=0.0028 err="request remained blocked after challenge orchestration retries" |
| target-home | https://www.target.com | status=200 blocked=true attempts=1 ms=19605 cost=0.0020 err="" | status=200 blocked=true attempts=3 ms=18952 cost=0.0113 err="request remained blocked after challenge orchestration retries" |
| target-search-page2 | https://www.target.com/s?searchTerm=laptop&page=2 | status=200 blocked=true attempts=1 ms=17087 cost=0.0026 err="" | status=200 blocked=true attempts=3 ms=20763 cost=0.0114 err="request remained blocked after challenge orchestration retries" |

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

