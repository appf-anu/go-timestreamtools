#/bin/bash

WORKING=`openssl rand -base64 16`
TMP="/short/xe2/gdd801/tmp/$WORKING"

./tsselect/./tsselect -source "$1" -start {{start}} -end {{end}} | \
 ./tsalign/./tsalign -output "$TMP" | \
 ./tsorganize/./tsorganize -del -output "$TMP"
