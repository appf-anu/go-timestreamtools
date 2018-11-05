#!/bin/bash
#qsub create_movie_single.pbs -q express -v START=2018-04-27,TRIAL=TR0117,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam01


# qsub create_movie_single.pbs -q express -N mov_120-Eu02-Cam01 -v START=2018-07-16,END=2018-08-23,STARTTOD="09:30",ENDTOD="16:30",TRIAL=TR0120,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam01
# qsub create_movie_single.pbs -q express -N mov_120-Eu02-Cam02 -v START=2018-07-16,END=2018-08-23,STARTTOD="09:30",ENDTOD="16:30",TRIAL=TR0120,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam02
# qsub create_movie_single.pbs -q express -N mov_120-Eu02-Picam -v START=2018-07-16,END=2018-08-23,STARTTOD="09:30",ENDTOD="16:30",TRIAL=TR0120,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Picam
#
# qsub create_movie_single.pbs -N mov_116-Eu02-Picam -v STARTTOD="06:30",ENDTOD="17:30",TRIAL=TR0116,START=2018-04-09,END=2018-04-27,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Picam
# qsub create_movie_single.pbs -N mov_116-Eu02-Cam01 -v STARTTOD="06:30",ENDTOD="17:30",TRIAL=TR0116,START=2018-04-09,END=2018-04-27,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam01
#
# qsub create_movie_single.pbs -N mov_117-Eu02-Picam -v STARTTOD="06:30",ENDTOD="17:30",START=2018-04-27,END=2018-07-06,TRIAL=TR0117,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Picam
# qsub create_movie_single.pbs -N mov_117-Eu02-Cam01 -v STARTTOD="06:30",ENDTOD="17:30",START=2018-04-27,END=2018-07-06,TRIAL=TR0117,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam01

qsub create_movie_single.pbs -N mov_120-Eu02-Picam -v STARTTOD="06:30",ENDTOD="17:30",TRIAL=TR0120,START=2018-07-16,END=2018-08-24,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Picam
qsub create_movie_single.pbs -N mov_120-Eu02-Cam01 -v STARTTOD="06:30",ENDTOD="17:30",TRIAL=TR0120,START=2018-07-16,END=2018-08-24,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam01

qsub create_movie_single.pbs -N mov_12-Eu02-Picam -v STARTTOD="06:30",ENDTOD="17:30",TRIAL=TR0122,START=2018-09-03,END=2018-10-02,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Picam
qsub create_movie_single.pbs -N mov_12-Eu02-Cam01 -v STARTTOD="06:30",ENDTOD="17:30",TRIAL=TR0122,START=2018-09-03,END=2018-10-02,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam01
qsub create_movie_single.pbs -N mov_12-Eu02-Cam02 -v STARTTOD="06:30",ENDTOD="17:30",TRIAL=TR0122,START=2018-09-03,END=2018-10-02,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam02

# qsub create_movie_single.pbs -N mov_119-GC35-Picam -v STARTTOD="06:30",ENDTOD="17:30",START=2018-04-27,TRIAL=TR0119,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC35-Picam
# qsub create_movie_single.pbs -N mov_119-GC35L -v STARTTOD="06:30",ENDTOD="17:30",START=2018-04-27,TRIAL=TR0119,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC35L
# qsub create_movie_single.pbs -N mov_119-GC35R -v STARTTOD="06:30",ENDTOD="17:30",START=2018-04-27,TRIAL=TR0119,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC35R

# qsub create_movie_single.pbs -N mov_116-Eu02-Picam -a 0100 -v START=2018-04-01,END=2018-04-27,TRIAL=TR0116,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Picam
# qsub create_movie_single.pbs -N mov_116-Eu02-Cam01 -a 0100 -v START=2018-04-01,END=2018-04-27,TRIAL=TR0116,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam01
# qsub create_movie_single.pbs -N mov_115-GC05L -a 0100 -v START=2018-04-01,END=2018-04-27,TRIAL=TR0115,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC05L,EXTRA=RGB01,STARTTOD="05:20",ENDTOD="16:40"
# qsub create_movie_single.pbs -N mov_115-GC05R -a 0100 -v START=2018-04-01,END=2018-04-27,TRIAL=TR0115,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC05R,EXTRA=RGB02,STARTTOD="05:20",ENDTOD="16:40"
# qsub create_movie_single.pbs -N mov_115-GC05-Picam -a 0100 -v START=2018-04-01,END=2018-04-27,TRIAL=TR0115,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC05-Picam,STARTTOD="05:20",ENDTOD="16:40"
# qsub create_movie_single.pbs -N mov_114-GC03R -a 0100 -v START=2018-04-01,END=2018-04-27,TRIAL=TR0114,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC03R,EXTRA=RGB02,STARTTOD="05:20",ENDTOD="16:40"
# qsub create_movie_single.pbs -N mov_114-GC03L -a 0100 -v START=2018-04-01,END=2018-04-27,TRIAL=TR0114,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC03L,EXTRA=RGB01,STARTTOD="05:20",ENDTOD="16:40"
# qsub create_movie_single.pbs -N mov_114-GC03-Picam -a 0100 -v START=2018-04-01,END=2018-04-27,TRIAL=TR0114,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC03-Picam,STARTTOD="05:20",ENDTOD="16:40"

#qsub create_movie_single.pbs -N "mov_TimHouse" -v NAME="TimHouse-SpringCam01",TRIAL=Misc,SOURCE=/home/801/gdd801/phenomics/camupload/picam/TimHouse-SpringCam01
#qsub create_movie_single.pbs -N "mov_Makerspace" -v NAME="MakerSpace-Picam01",TRIAL=Misc,SOURCE=/home/801/gdd801/phenomics/camupload/picam/MakerspacePicam01
#qsub create_movie_single.pbs -N "mov_BVZ-House" -v NAME="BVZ-HousePicam",TRIAL=Misc,SOURCE=/home/801/gdd801/phenomics/camupload/picam/BVZ-HousePicam

# qsub create_movie_single.pbs -N mov_123-GC37-Picam -v STARTTOD="05:00",ENDTOD="17:00",START=2018-08-01,TRIAL=TR0123,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC37-Picam
# qsub create_movie_single.pbs -N mov_123-GC37L -v STARTTOD="05:00",ENDTOD="17:00",START=2018-08-01,TRIAL=TR0123,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC37L
# qsub create_movie_single.pbs -N mov_123-GC37R -v STARTTOD="05:00",ENDTOD="17:00",START=2018-08-01,TRIAL=TR0123,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC37R
# qsub create_movie_single.pbs -N mov_123-GC37LN -v STARTTOD="05:00",ENDTOD="17:00",START=2018-08-01,TRIAL=TR0123,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC37L-NIR-C01
# qsub create_movie_single.pbs -N mov_123-GC37RN -v STARTTOD="05:00",ENDTOD="17:00",START=2018-08-01,TRIAL=TR0123,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC37R-NIR-C02
