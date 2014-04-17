# goflake

goflake is an implementation of [Twitter's Snowflake](https://github.com/twitter/snowflake/) in Go based on [sdming/gosnow](https://github.com/sdming/gosnow) and [davegardnerisme/cruftflake](https://github.com/davegardnerisme/cruftflake).

goflake can be used to generate unique 64bit IDs without coordination, these consist of:

 * time - 41 bits (millisecond precision w/ a custom epoch of 2012-01-01 00:00:00 gives us 69 years)
 * configured worker id - 10 bits - gives us up to 1024 workers
 * sequence number - 12 bits - rolls over every 4096 per worker (with protection to avoid rollover in the same ms)

## Usage

```golang
package main

import (
    "github.com/mattheath/goflake"
    "fmt"
)

func main() {

    // Create a new goflake with a worker id, these must be unique.
    v, err := goflake.New(100)

    for i := 0; i < 10; i++ {
        id, err := v.Generate()
        fmt.Println(id)
    }
}
```
