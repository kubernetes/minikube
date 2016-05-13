#!/bin/bash

set -e

./autogen.sh && ./configure && make
