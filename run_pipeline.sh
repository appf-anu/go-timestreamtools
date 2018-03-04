#/bin/bash

WORKING=`openssl rand -base64 16`
TMP="/short/xe2/gdd801/tmp/$WORKING"

./tsselect/./tsselect -source "$1" -start {{start}} -end {{end}} | \
 ./tsalign/./tsalign -output "$TMP" | \
 ./tsorganize/./tsorganize -del -output "$TMP" | \
 ./tsrename/./tsrename -del -name "{{name}}~fullres" | \
 ./tsresize/./tsresize -res 1920x1080 | \
 ./tsrename/./tsrename -del -name "{{name}}~1920"
