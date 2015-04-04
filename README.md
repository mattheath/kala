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
    "github.com/mattheath/kala/snowflake"
    "fmt"
)

func main() {

    // Create a new snowflake compatible minter with a worker id, these *must* be unique.
    v, err := snowflake.New(100)

    for i := 0; i < 10; i++ {
        id, err := v.Mint()
        fmt.Println(id)
    }
}
```

## Bigflake

Kāla provides an alternative minter which mints larger 128bit ids,
in a similar way to Boundary's [Flake](https://github.com/boundary/flake) implementation. These consist of:

 * time - 64bits
 * worker ID - 48 bits
 * sequence number - 16 bits

Example IDs:
```
26341991268378369512474991263745
26341991268378369512474991263746
26341991268378369512474991263747
26341991268378369512474991263748
26341991268378369512474991263749
26341991268378369512474991263750
26341991268378369512474991263751
26341991268378369512474991263752
26341991268378369512474991263753
26341991268378369512474991263754
```

### Usage

```golang
package main

import (
    "github.com/mattheath/kala/bigflake"
    "github.com/mattheath/kala/util"
    "fmt"
)

func main() {
	// Using mac address as worker id
	mac := "80:36:bc:db:64:16"
	workerId, err := util.MacAddressToWorkerId(mac)

    // Create a new bigflake minter with a worker id
    m, err := bigflake.New(workerId)

    for i := 0; i < 10; i++ {
        id, err := m.Mint()
        fmt.Println(id)
    }
}
```

## Benchmarks

Implementations are reasonably fast, but will of course vary depending on hardware. The below are from a 1.7Ghz i7 Macbook Air:

```
BenchmarkMintBigflakeId   2000000        1.952s        976 ns/op
BenchmarkMintSnowflakeId  5000000        1.575s        315 ns/op
```
