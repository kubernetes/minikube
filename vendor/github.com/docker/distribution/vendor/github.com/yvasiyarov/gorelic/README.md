# GoRelic

New Relic agent for Go runtime. It collect a lot of metrics about scheduler, garbage collector and memory allocator and 
send them to NewRelic.

### Requirements  
- Go 1.1 or higher
- github.com/yvasiyarov/gorelic
- github.com/yvasiyarov/newrelic_platform_go
- github.com/yvasiyarov/go-metrics

You have to install manually only first two dependencies. All other dependencies will be installed automatically 
by Go toolchain.   

### Installation   
```bash
go get github.com/yvasiyarov/gorelic
```
and add to the initialization part of your application following code:  
```go
import (
    "github.com/yvasiyarov/gorelic"
)
....

agent := gorelic.NewAgent()
agent.Verbose = true
agent.NewrelicLicense = "YOUR NEWRELIC LICENSE KEY THERE"
agent.Run()

```

### Middleware  
If you using Beego, Martini, Revel or Gin framework you can hook up gorelic with your application by using the following middleware:
- https://github.com/yvasiyarov/beego_gorelic   
- https://github.com/yvasiyarov/martini_gorelic   
- https://github.com/yvasiyarov/gocraft_gorelic   
- http://wiki.colar.net/revel_newelic
- https://github.com/jingweno/negroni-gorelic
- https://github.com/brandfolder/gin-gorelic
   

### Configuration  
- NewrelicLicense - its the only mandatory setting of this agent.
- NewrelicName - component name in NewRelic dashboard. Default value: "Go daemon"
- NewrelicPollInterval - how often metrics will be sent to NewRelic. Default value: 60 seconds
- Verbose - print some usefull for debugging information. Default value: false
- CollectGcStat - should agent collect garbage collector statistic or not. Default value: true
- CollectHTTPStat - should agent collect HTTP metrics. Default value: false
- CollectMemoryStat - should agent collect memory allocator statistic or not. Default value: true
- GCPollInterval - how often should GC statistic collected. Default value: 10 seconds. It has performance impact. For more information, please, see metrics documentation.
- MemoryAllocatorPollInterval - how often should memory allocator statistic collected. Default value: 60 seconds. It has performance impact. For more information, please, read metrics documentation.


## Metrics reported by plugin
This agent use functions exposed by runtime or runtime/debug packages to collect most important information about Go runtime.

### General metrics   
- Runtime/General/NOGoroutines - number of runned go routines, as it reported by NumGoroutine() from runtime package
- Runtime/General/NOCgoCalls - number of runned cgo calls, as it reported by NumCgoCall() from runtime package

### Garbage collector metrics      
- Runtime/GC/NumberOfGCCalls - Nuber of GC calls, as it reported by ReadGCStats() from runtime/debug 
- Runtime/GC/PauseTotalTime - Total pause time diring GC calls, as it reported by ReadGCStats() from runtime/debug (in nanoseconds)
- Runtime/GC/GCTime/Max - max GC time
- Runtime/GC/GCTime/Min - min GC time
- Runtime/GC/GCTime/Mean - GC mean time
- Runtime/GC/GCTime/Percentile95 - 95% percentile of GC time

All this metrics are measured in nanoseconds. Last 4 of them can be inaccurate if GC called more often then once in GCPollInterval. 
If in your workload GC is called more often - you can consider decreasing value of GCPollInterval. 
But be carefull, ReadGCStats() blocks mheap, so its not good idea to set GCPollInterval to very low values.

### Memory allocator 
- Component/Runtime/Memory/SysMem/Total - number of bytes/minute allocated from OS totally. 
- Component/Runtime/Memory/SysMem/Stack - number of bytes/minute allocated from OS for stacks.
- Component/Runtime/Memory/SysMem/MSpan - number of bytes/minute allocated from OS for internal MSpan structs.
- Component/Runtime/Memory/SysMem/MCache - number of bytes/minute allocated from OS for internal MCache structs.
- Component/Runtime/Memory/SysMem/Heap - number of bytes/minute allocated from OS for heap.
- Component/Runtime/Memory/SysMem/BuckHash - number of bytes/minute allocated from OS for internal BuckHash structs.
- Component/Runtime/Memory/Operations/NoFrees - number of memory frees per minute
- Component/Runtime/Memory/Operations/NoMallocs - number of memory allocations per minute
- Component/Runtime/Memory/Operations/NoPointerLookups - number of pointer lookups per minute
- Component/Runtime/Memory/InUse/Total - total amount of memory in use
- Component/Runtime/Memory/InUse/Heap - amount of memory in use for heap
- Component/Runtime/Memory/InUse/MCacheInuse - amount of memory in use for MCache internal structures
- Component/Runtime/Memory/InUse/MSpanInuse - amount of memory in use for MSpan internal structures  
- Component/Runtime/Memory/InUse/Stack - amount of memory in use for stacks

### Process metrics
- Component/Runtime/System/Threads - number of OS threads used
- Runtime/System/FDSize - number of file descriptors, used by process
- Runtime/System/Memory/VmPeakSize - VM max size
- Runtime/System/Memory/VmCurrent  - VM current size
- Runtime/System/Memory/RssPeak    - max size of resident memory set
- Runtime/System/Memory/RssCurrent - current size of resident memory set   

All this metrics collected once in MemoryAllocatorPollInterval. In order to collect this statistic agent use ReadMemStats() routine.
This routine calls stoptheworld() internally and it block everything. So, please, consider this when you change MemoryAllocatorPollInterval value.

### HTTP metrics   
- throughput (requests per second), calculated for last minute  
- mean throughput (requests per second)   
- mean response time  
- min response time  
- max response time  
- 75%, 90%, 95% percentiles for response time
 

In order to collect HTTP metrics, handler functions must be wrapped using WrapHTTPHandlerFunc:

```go
http.HandleFunc("/", agent.WrapHTTPHandlerFunc(handler))
```

## TODO
- Collect per-size allocation statistic
- Collect user defined metrics

