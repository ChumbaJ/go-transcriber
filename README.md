# Structure

Input - Queue - Output

## Input:

1. Upload audio-file
2. Worker slices audio file into chunks:
   Each chunk is saved into storage
   Saved chunk is indexed by order
3. Job with all chunks addrses is saved into database

  Job is only ready if it has all chunks of audio-file stored and indexed and has links to then in storage
  
  The structure of job:
  
  ```
    {
      id
      chunks: [
        {
          order
          addr
        },
        {
          ...
        },
      ]
      transcribed: {
        1: "..."
        2: "..."
        3: "..."
      }
    }
  ```
```
```



## Queue:

  The structure of queue item:

  ```
    {
      jobId
      chunk: {
        order
        addr
      }
    }
  ```
```
```


## Output:

1. Takes queue item from queue
2. Gets file chunk from storage and send to LLM
3. When LLM returns transcribed value (as text) we save the text into database
4. The transcribed text is saved into job.transcribed[order] map, therefore each text part is now ordered in map.
   And we can build the output easily
5. At the moment when worker saves chunk, it checks if all chunks were transcribed and if yes then job is marked as "done"





