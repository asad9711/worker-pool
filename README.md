# worker-pool
Implement Go concurrency pattern - worker pool


## Idea
- Use go concurrency model to open a number of workers and provide them tasks to process.
- Program gets webpages, extracts content length of a page and puts this value into the map.
When all address processed, program prints all pairs (webpage address, content length).

### Specific requirements:
- When all webpages from webPages slice processed, print each key-value from webPages.
    
    Example:

    
    google.com - 4501  
    ...

### NOTE:
- The program's execution time should be configurable.
- If execution time is out, the program should gracefully shut all currently running workers.