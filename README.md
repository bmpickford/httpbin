# HTTPBin
Small CLI application to quickly spin up a server to collect requests and dump them to a har file or stdout for debugging without setting up a proxy

## Why would I use this
 * Easily print local requests
 * Easily test against different status codes
 * Generate har from local requests

## Demo
![ezgif com-gif-maker](https://user-images.githubusercontent.com/12684286/120925929-23e3a580-c71e-11eb-9f38-1e9d8e6253dc.gif)



## Installation

1. Download from [releases](https://github.com/bmpickford/httpbin/releases/latest)
1. `tar -zxvf httpbin_X.X.X_*.tar.gz`
1. `mv httpbin_X.X.X_*/httpbin /usr/local/bin`


## Usage

`httpbin [OPTIONS]...`

### Flags
| Flag | Type    | Default | Description |
| ---- |:-------:|:-------:|:-----|
| port | Number  | 8080    | Port number to run on |
| out  | String  | stdout  | Name of outfile. Will use stdout if nothing is supplied |
| har  | Boolean | false   | Use har format |


### Example
Output request on port 5000 to a har file named myreqs.har

`httpbin -port 5000 -har -out myreqs`


## Use Cases
 > As a developer, I don't have an API ready yet but would like to output some requests to inspect the data

 > As a developer, I would like to manually see how my System would respond to different HTTP response codes for a specific request

 > As a developer, I would like to debug my application locally by dumping all my request information being sent

TODO: tests
