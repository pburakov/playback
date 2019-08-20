# playback

CLI tool for "replaying" events from local file into PubSub topic.

## Description

Playback tool provides flexible ways of replaying a data stream into a PubSub topic packaged in a single binary. Intended and typical use cases for the Playback tool include:
 - A/B testing of streaming pipelines and data ingestion services;
 - streaming of data dumped from BigQuery/SQL tables into Pub/Sub without using Dataflow;
 - real time playback of [Dataflow streams](https://console.cloud.google.com/dataflow/createjob) dumped to GCS using the "Cloud Pub/Sub to Text Files on Cloud Storage" template.

## Installation

On macOS, playback tool can be installed using Homebrew:

```bash
$ brew tap pburakov/io
$ brew install pburakov/io/playback
```

## Build

Alternatively, with Go version 1.11.4 (or greater) installed, you can build the binary from the root of this repository by running:

```bash
$ go build
```

Go will download the dependencies and compile `playback` binary which can be ran from shell using `./playback`.

## Usage

Basic usage: 

```
$ playback -input=<input_file> -ts_column=<ts_column> -project_id=<output_gcp_project> -topic=<pubsub_topic> [args...] 
```

Advanced example:

```bash
$ playback -mode=2 -window=1000 -input=data.json -ts_column=created_at -project_id=my-project -topic=my-topic 
``` 

To access default values and detailed info on program arguments in your shell, run:  

```bash
$ playback --help
```

## Playback Modes

Playback tool provides 3 modes of operation: paced (default), instant and relative. 

- In **relative mode**, the relative distance between two consecutive event timestamps is closely maintained. This mode is useful for emulating or replaying real-time traffic. Relative mode is comparatively more expensive, since the input row has to be first parsed and searched for the timestamp. For predictable results, the input data must be sorted by the timestamp column, defined as a program argument (see [Settings](#settings)).

- In **paced mode**, messages are played back one by one at configurable equal intervals with the original event timestamp being ignored. Paced mode is useful for limiting throughput and maintaining order of events in the output.

- In **instant mode**, the input data is sent immediately with the original event timestamp being ignored. This is the most resource-demanding mode of operation, recommended only when the total number of events is relatively small. Consider using [Dataflow template](https://console.cloud.google.com/dataflow/createjob) named "Text Files Cloud Storage to Cloud Pub/Sub" as a scalable alternative.

It is important to note that all modes (and instant mode is the most vulnerable) are subject to IO constraints, CPU, available memory, event payload size and network throughput. Throttling is not implemented. It is not guaranteed that outgoing messages will reach PubSub at the specified timestamp, or in the specified order.

## Settings

| Flag | Type | Mode | Required | Description |
|------|------|------|----------|-------------|
| `mode` | int | - | false | Playback mode: `0` - paced (default), `1` - instant, and `2` - relative. |
| `input` | string | all | true | Path to the input file. Supported formats: JSON (newline delimited), CSV and Avro.
| `project_id` | string | all | true | Output Google Cloud project id. |
| `topic` | string | all | true | Output PubSub topic. |
| `ts_column` | string | Relative | true | Name of the timestamp column for relative playback mode. The input data must be sorted by that column. |
| `ts_format` | string | Relative | false | Timestamp format for relative playback mode. Layouts must use the reference time Mon Jan 2 15:04:05 MST 2006 to show the pattern with which to parse a given string. Refer to this [documentation](https://golang.org/pkg/time/#pkg-constants) for more detail. |
| `delay` | int | Paced | false | Delay between line reads for paced playback, in milliseconds. | 
| `window` | int | all | false | Event accumulation window for relative playback mode, in milliseconds. Use higher values if input event distribution on the timeline is sparse, lower values for a more dense event distribution. |
| `jitter` | int | all | false | Max jitter for relative and paced playback modes, in milliseconds. | 
| `timeout` | int | all | false | Publish request timeout, in milliseconds. |

## Known Bugs and Limitations

- Using timestamp field within a nested structure is not currently supported.
- Last message in JSON files will be skipped if there's no newline delimiter at the EOF. 

## Supported Formats

Playback tool supports JSON (newline delimited), CSV and Avro files, typically produced and consumed by Google BigQuery and Dataflow stack.

JSON and Avro formats guarantee schema compliance and support for nested structures. While JSON and Avro events are published as is (byte-wise), CSV data is converted to JSON key-value object with strings as keys and values.
