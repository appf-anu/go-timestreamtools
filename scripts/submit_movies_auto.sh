#!/bin/bash
qsub create_movie.pbs -v START=2018-04-27,TRIAL=TR0117,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Picam
qsub create_movie.pbs -v START=2018-04-27,TRIAL=TR0117,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam01
# qsub create_movie.pbs -v START=2018-04-01,TRIAL=TR0115,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC05L,EXTRA=RGB01,STARTTOD="05:20",ENDTOD="16:40"
# qsub create_movie.pbs -v START=2018-04-01,TRIAL=TR0115,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC05R,EXTRA=RGB02,STARTTOD="05:20",ENDTOD="16:40"
# qsub create_movie.pbs -v START=2018-04-01,TRIAL=TR0115,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC05-Picam,STARTTOD="05:20",ENDTOD="16:40"
# qsub create_movie.pbs -v START=2018-04-01,TRIAL=TR0114,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC03R,EXTRA=RGB02,STARTTOD="05:20",ENDTOD="16:40"
# qsub create_movie.pbs -v START=2018-04-01,TRIAL=TR0114,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC03L,EXTRA=RGB01,STARTTOD="05:20",ENDTOD="16:40"
# qsub create_movie.pbs -v START=2018-04-01,TRIAL=TR0114,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC03-Picam,STARTTOD="05:20",ENDTOD="16:40"
