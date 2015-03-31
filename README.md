# Kāla

[![Build Status](https://travis-ci.org/mattheath/goflake.svg?branch=master)](https://travis-ci.org/mattheath/kala)

Kāla provides implementations of time ordered distributed ID generators in Go.

Snowflake: Generates 64bit k-ordered IDs similar to [Twitter's Snowflake](https://github.com/twitter/snowflake/) based on [sdming/gosnow](https://github.com/sdming/gosnow) and [davegardnerisme/cruftflake](https://github.com/davegardnerisme/cruftflake). (Formerly [github.com/mattheath/goflake](https://github.com/mattheath/goflake))

BigFlake: Generates 128bit k-ordered IDs based on Boundary's [Flake](https://github.com/boundary/flake) implementation, either with or without coordination.

## Snowflake

Kāla's snowflake compatible minter can be used to generate unique 64bit IDs without coordination, these consist of:

 * time - 41 bits (millisecond precision w/ a custom epoch of 2012-01-01 00:00:00 gives us 69 years)
 * configured worker id - 10 bits - gives us up to 1024 workers
 * sequence number - 12 bits - rolls over every 4096 per worker (with protection to avoid rollover in the same ms)

Example IDs:
```
429587937416445952
429587937416445953
429587937416445954
429587937416445955
429587937416445956
429587937416445957
429587937416445958
429587937416445959
429587937416445960
429587937416445961
```

### Usage

```golang
package main

import (
    "github.com/mattheath/kala"
    "fmt"
)

func main() {

    // Create a new snowflake compatible minter with a worker id, these *must* be unique.
    v, err := kala.NewSnowflake(100)

    for i := 0; i < 10; i++ {
        id, err := v.Mint()
        fmt.Println(id)
    }
}
```
