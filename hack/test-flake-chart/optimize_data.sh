#!/bin/bash

awk -F, 'BEGIN {OFS = FS} { for(i=1; i<=NF; i++) { if($i == j[i]) { $i = ""; } else { j[i] = $i; } } printf "%s\n",$0 }'
