# worker-pattern

Golang Worker Pattern w/ Redis

## Requirements

To build & run the project you will need:
* [go](https://golang.org/)  
* [make](https://www.gnu.org/software/make/)  
* [docker](https://docs.docker.com/)  
* [docker-compose](https://docs.docker.com/compose/)  

## Building & Running

To build & run the project simply run `make`:
```
$ make
```

## Testing

To run the unit tests you can run:
```
$ make test
```

This will run the unit tests mocked against the Redis client that wraps the
required commands for the interactions with Redis. There are no integration
tests as I did not get around to writing them however when running in
docker-compose it does correctly execute against the redis container

## Usage Instructions

When run the program will populate Redis with a set number of iterations that
can be toggled via the `--iterations` flag. This simply sets a command of
`sleep` into the cache along with a randomly generated delay (between 10-60ms
although it is using math/rand - not super secure). Once the work has been
populated into the Redis instance the program will then spawn workers to act on
it. The workers will pull the subscribe and retrieve the data from Redis and
then finish once there is no iterations left. It will then display the total
amount of time each worker slept for. The number of workers can be toggled using
the `--workers` flag. These flags are both configurable via the
`docker-compose.yaml` file.

## Scaling

While this service can be scaled horizontally; the workers only process a single
key at a time. This was due to time constraints & the fact it is only using a
stub for a work item (sleep). A better approach would be to allow each worker to
process more than one item using a go routine concurrently at have the status
put back onto the queue.

## Sequential vs Parallel tasks?

The work being done in this program is task being completed in parallel as the
workers do not need any context as to the other work being processed by the
others. These type of tasks are certainly useful when you want to complete one
task faster by committing more workers to it. This was certainly the ideal
approach for this task given all of the subtasks are identical and can be
assigned to different workers for a much faster execution cycle.

There maybe however, times when work is needed to be processed one after the
other. Redis supports a few ways of doing this. First we could use transactions
which allow commands to be executed sequentially and have a worker process this
information. The commands in a transaction are serialized and executed
sequentially. It can never happen that a request issued by another client is
served in the middle of the execution of a Redis transaction. This guarantees
that the commands are executed as a single isolated operation.

Another option would be to use Redis streams which allows for following an
append only data stream. As a stream of messages that can be partitioned to
multiple consumers that are processing such messages, so that groups of
consumers can only see a subset of the messages arriving in a single stream.
This allows workers to either specifically process their own work much like
Kafka consumers with "topics".

Finally as Redis allows any data to be inserted via hashes the worker process
could simply put information back onto the queue that signals to another worker
and/or different process that sequential work is needed after the first task has
been processed. These workers can then retry and backoff exponentially until
needed and validated that all of the required dependencies had been met before
moving onto the next task

// expected in a failure scenario
Job A (completed)-> Job B (failed) -> Job C (not executed)

// expected in the ideal scenario
Job A (completed)-> Job B (completed) -> Job C (completed)

## Shortcomings of the code?

Apart from the code being written in a few hours there are several
shortcomings. First being that there is no feedback into the system that the
work has not been completed. This is simply logged, dropped and the worker moves
on. This could be improved by having a retry loop around the processing of the
information. This could allow the worker to avoid any small network blips or
other little interruptions to the connection that could occur. The program could
also put feedback back into the queue when a task could not be completed
(although this does become an issue in the failure scenario of losing the
connection). This would allow other processes to check the failed queue to see
if the work could be reprocessed at a later date.

Secondly Time-outs could be appended to the hashes in Redis to ensure that a
task is retried at the earliest convenience. Another goroutine or process could
check to ensure that messages with a higher priority (I.E sat longer in the
queue due to a failed run) could be processed as soon as the next worker is
available.

Finally as mentioned above in the scaling piece the workers are only processing
one item off the queue at a time. This could cause a bottleneck within the
system if all of the workers were currently processing large workloads. The
overhead could quickly become apparent as the work could easily be done
concurrently.
