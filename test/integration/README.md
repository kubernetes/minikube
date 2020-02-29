# Integration tests

## The basics

To run all tests from the minikube root directory:

`make integration`

## Quickly iterating on a single test

Run a single test on an active cluster:

`make integration -e TEST_ARGS="-test.run TestFunctional/parallel/MountCmd --profile=minikube --cleanup=false"`

WARNING: For this to work repeatedly, the test must be written so that it cleans up after itself.

See `main.go` for details.

## Run integration test for a specific driver in parallel with HTML report

for html report install gopogh https://github.com/medyagh/gopogh/releases/latest

```
make integration -e TEST_ARGS="-minikube-start-args=--vm-driver=docker"  2>&1 | tee ./out/testout.txt

```

```
go tool test2json -t < ./out/testout.txt > ./out/testout.json || true
medya@~/workspace/minikube (gopogh_kic) $ gopogh_status=$(gopogh -in "./out/testout.json" -out "./out/testout.html" -name "whatver" -pr "whatever" -repo github.com/kubernetes/minikube/  -details "${COMMIT}") || true
```


## Disabling parallelism

`make integration -e TEST_ARGS="-test.parallel=1"`

## Testing philosophy

- Tests should be so simple as to be correct by inspection
- Readers should need to read only the test body to understand the test
- Top-to-bottom readability is more important than code de-duplication

Tests are typically read with a great air of skepticism, because chances are they are being read only when things are broken. 
