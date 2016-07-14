# Gannett's Realtime Analytics [ETL](https://en.wikipedia.org/wiki/Extract,_transform,_load)

* Extract data from Chartbeat's API
* Transform it into meaningful data we can use for realtime analytics
* Load into a mongo database

## Setup

1. Install Go dependencies:

    ```bash
    go get
    ```

2. Build or install the Go binary:

    ```bash
    go build
    ```

    -or-

    ```bash
    go install
    ```

3. Set up environment variables.

    It is recommended that you copy `.env.sample` into `.env` and adjust as necessary.
    Apply them via `source .env`

4. Run `chartbeat-api` from `$GOPATH/bin/`

## [Modd](https://github.com/cortesi/modd)

Modd will reinstall the application and run it again whenever there is a filesystem change in the
source application directory.  Simply install `modd` then run

```bash
modd
```

## Command options

```bash
chartbeat-api help

Usage of ./chartbeat-api:
  -l int
    Time in seconds to sleep before looping and hitting the apis again
```

## Required environment variables

```
DOMAINS=this.com,that.com
CHARTBEAT_API_KEY=your-key-here
MONGO_URI=mongodb://127.0.0.1:27017/chartbeat-api
```
