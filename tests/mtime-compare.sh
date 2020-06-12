#!/bin/bash

DIR_A=$1
DIR_B=$2

export RESULT=0

function onerror() {
    echo $1
    date -r $1
    date -r ${DIR_B}/$1
    RESULT = 1
    export RESULT
}

cd ${DIR_A}
for file in `\find . -type f`; do
    if [ $file -nt ${DIR_B}/${file} ]; then
        echo $file
    fi
    if [ $file -ot ${DIR_B}/${file} ]; then
        echo $file
    fi
done
cd -

exit ${RESULT}
