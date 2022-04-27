### USAGE ###
## DOWNLOAD METRICS ##
- curl -o ./heap.out http://localhost:9090/pprof/heap
- curl -o ./profile.out http://localhost:9090/pprof/profile?seconds=5
- curl -o ./block.out http://localhost:9090/pprof/block
- curl -o ./mutex.out http://localhost:9090/pprof/mutex
- curl -o ./mutex.out http://localhost:9090/pprof/trace

## OPEN METRICS ##
- go tool pprof ./heap.out
- go tool pprof ./profile.out
- go tool pprof ./block.out
- go tool pprof ./mutex.out
- go tool trace ./trace.out

### PPROF OPTIONS ###
- top
- web
- list (func name)
