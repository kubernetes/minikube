# Integration tests

## The basics

To run all tests from the minikube root directory:

`make integration`

## Quickly iterating on a single test

Run a single test on an active cluster:

`make integration -e TEST_ARGS="-test.v -test.run TestFunctional/parallel/MountCmd --profile=minikube --cleanup=false"`

WARNING: For this to work repeatedly, the test must be written so that it cleans up after itself.

See `main.go` for details.

## Disabling parallelism

`make integration -e TEST_ARGS="-test.parallel=1"`

## Testing philosophy

