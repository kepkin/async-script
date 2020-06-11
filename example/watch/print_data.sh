#!/bin/bash

letters=(a b c d e f g h j k l)

for i in `seq 0 10`; do
  for k in `seq 0 $i`; do
    echo ${letters[$i]} $k
  done
  sleep 1
done