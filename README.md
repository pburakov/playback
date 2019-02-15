# playback

CLI tool for "replaying" events from local file into PubSub topic.

## Installation

On macOS, playback tool can be installed using Homebrew:

```bash
$ brew tap pburakov/io
$ brew install pburakov/io/playback
```

## Build

With Go version 1.11.4 (or greater) installed, you can build the binary from the root of this repository by running:

```bash
$ go build
```

Go will download the dependencies and compile `playback` binary which can be ran from shell using `./playback`.

## Usage

```
$ playback -i <input_file> -c <ts_column> -p <output_gcp_project> -t <pubsub_topic> [params...] 
```

Example:

```bash
$ playback -m 2 -w 1000 -i data.json -c created_at -p my-project -t my-topic 
``` 

To access default values and detailed info on program parameters in your shell, run:  

```bash
$ playback --help
```

## Playback Modes

Playback tool provides 3 modes of operation: paced (default), instant and relative. 

- In **paced mode**, messages are played back one by one at configurable equal intervals. Original event timestamp is bypassed. Useful for limiting throughput and maintaining order of events in the output.

- In **relative mode** the relative difference between event timestamps is maintained. This mode is useful for emulating or replaying real-world traffic. Relative mode is comparatively more expensive, since the input row has to be first parsed and searched for the timestamp.   

- In **instant mode**, the input data is sent immediately without delay after read. Original event timestamp is bypassed. This is the most resource-demanding mode of operation, recommended only when the total number of events is relatively small. Consider using [Dataflow template](https://console.cloud.google.com/dataflow/createjob) named "Text Files Cloud Storage to Cloud Pub/Sub" as an alternative.

**Please note**, all modes (and instant mode is the most vulnerable) are subject to input / output constraints, CPU, available memory, event payload size and network throughput. Throttling is not implemented. It is not guaranteed that outgoing messages will reach PubSub at the specified timestamp, or in the specified order.

## Settings

| Flag | Type | Required | Description |
|------|------|----------|-------------|
| `i` | string | true | Path to input file. Supported formats: JSON (newline delimited), CSV and Avro.
| `p` | string | true | Output Google Cloud project id. |
| `t` | string | true | Output PubSub topic. |
| `m` | int | false | Playback mode: `0` - paced, `1` - instant, and `2` - relative. |
| `c` | string | false | Name of the timestamp column for relative playback mode. The input data must be sorted by that column. |
| `f` | string | false | Timestamp format for relative playback mode. Layouts must use the reference time Mon Jan 2 15:04:05 MST 2006 to show the pattern with which to format/parse a given time/string. Refer to this [documentation](https://golang.org/pkg/time/#pkg-constants) for more detail. |
| `d` | int | false | Delay between line reads for paced playback, in milliseconds. | 
| `w` | int | false | Event accumulation window for relative playback mode, in milliseconds. Use higher values if input event distribution on the timeline is sparse, lower values for a more dense event distribution. |
| `j` | int | false | Max jitter for relative and paced playback modes, in milliseconds. | 
| `o` | int | false | Publish request timeout, in milliseconds. |

_More detailed info to be added_

## Supported Formats

Playback tool supports JSON (newline delimited), CSV and Avro files, typically supported by Google BigQuery and Dataflow stack.

Intended and typical use case for the Playback tool include:
 - streaming data stored in BigQuery tables into Pub/Sub;
 - replaying [Dataflow streams](https://console.cloud.google.com/dataflow/createjob) set up using the "Cloud Pub/Sub to Text Files on Cloud Storage" template;
 - A/B testing for streaming pipelines and data injestion services.    

JSON and Avro formats guarantee schema compliance and support for nested structures. While JSON and Avro events are published as is (byte-wise), CSV data is converted to JSON key-value object with strings as keys and values.
