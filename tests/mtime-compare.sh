#!/bin/bash

DIR_A=$(cd $1; pwd)
DIR_B=$(cd $2; pwd)

export RESULT=0

function onerror() {
    echo =======================
    echo $1
    date -r $1
    date -r ${DIR_B}/$1
    echo =======================
    export RESULT=1
}

cd ${DIR_A}
for file in `\find . -type d -name .git -prune -o -type f`; do
    if [ $file -nt ${DIR_B}/${file} ]; then
        onerror $file
    fi
    if [ $file -ot ${DIR_B}/${file} ]; then
        onerror $file
    fi
done
cd -

exit ${RESULT}
