#/bin/bash
NAME=
RESOLUTION=
START=
END=

WORKING=`openssl rand -base64 16`
TMP="/short/xe2/gdd801/tmp/$WORKING"


./tsselect/./tsselect -source "$1" -start "$START" -end "$END" | \
 ./tsalign/./tsalign -output "$TMP" | \
 ./tsorganize/./tsorganize -del -output "$TMP" | \
 ./tsrename/./tsrename -del -name "$NAME~fullres" | \
 ./tsresize/./tsresize -res 1920x1080 | \
 ./tsrename/./tsrename -del -name "$NAME~1920"
