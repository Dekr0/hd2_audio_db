## Reduce memory footprint

- Memory will intensively ramp up when extracting raw binary content of a Soundbank and 
writing them into a file.
- The above ramp up rate will increase when the extraction task is operating under 
go coroutine mode.
- After extraction, those raw binary content are no longer useful. They should be 
recycle.
- To reduce memory ramp up rate, one solution is memory pool. Since the raw binary 
content of a Soundbank is written, it's no longer needed. Thus, they can be reused.
- This doesn't fix the issue where there are still left over memory usage after 
the extractting raw binary content of Soundbanks. GC doesn't immedately recycle 
them.

## Reduce memory footprint caused by Map

## Go coroutine Pool
