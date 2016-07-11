# Golang Chartbeat API (not really)

Fetch Chartbeat data for Gannett properties and store those in Mongo. Yeah, not an API at all, sorry.

## Setup

1. Install Go dependencies:

        go get

2. Build the Go binary:

        go install

3. Set up environment variables. It is recommended that you copy `.env.sample` into `.env` and adjust as necessary. Apply them via `source .env` or, better yet, use [autoenv](https://github.com/horosgrisa/autoenv).

4. Run `chartbeat-api` from `$GOPATH/bin/`.

During development, you can run and restart the server via modd:

    modd

## Required environment variables

```
DOMAINS=this.com,that.com
CHARTBEAT_API_KEY=your-key-here
MONGO_URI=mongodb://127.0.0.1:27017/chartbeat-api
```
