#!/bin/bash
while read f
do
    if ffprobe -loglevel warning $f
    then
        printf "file '%s'\n" $f
    fi
done < $1
