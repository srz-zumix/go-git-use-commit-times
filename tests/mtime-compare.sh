#!/bin/bash

DIR_A=$1
DIR_B=$2

cd ${DIR_A}
for file in `\find . -type f`; do
    if [ $file -nt ${DIR_B}/${file} ]; then
        echo $file
        exit 1
    fi
    if [ $file -ot ${DIR_B}/${file} ]; then
        echo $file
        exit 1
    fi
done
