#!/bin/bash

RESOLUTION_HIGH_PICAM=3280x2464
RESOLUTION_HIGH_DSLR=5184x3456
RESOLUTION_PICAM=1920x1442
RESOLUTION_DSLR=1920x1280

# qsub run_pipeline_single.pbs -N "proc_TimHouse" -v NAME="TimHouse-SpringCam01",INTERVAL="1m",TRIAL=Misc,SOURCE=/home/801/gdd801/phenomics/camupload/picam/TimHouse-SpringCam01
# qsub run_pipeline_single.pbs -a 0100 -N "proc_Makerspace" -v NAME="MakerSpace-Picam01",TRIAL=Misc,SOURCE=/home/801/gdd801/phenomics/camupload/picam/MakerspacePicam01
# qsub run_pipeline_single.pbs -a 0100 -N "proc_BVZ-House" -v NAME="BVZ-HousePicam",TRIAL=Misc,SOURCE=/home/801/gdd801/phenomics/camupload/picam/BVZ-HousePicam
#
# qsub run_pipeline_single.pbs -a 0100 -N "proc_114_03L" -v TRIAL=TR0114,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC03R,EXTRA=RGB01,START=2018-04-01,END=2018-04-26,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_114_03R" -v TRIAL=TR0114,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC03L,EXTRA=RGB02,START=2018-04-01,END=2018-04-26,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_114_03P" -v TRIAL=TR0114,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC03-Picam,START=2018-04-01,END=2018-04-26,RESOLUTION_HIRES="$RESOLUTION_HIGH_PICAM",RESOLUTION="$RESOLUTION_PICAM"
#
# qsub run_pipeline_single.pbs -a 0100 -N "proc_115_05L" -v TRIAL=TR0115,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC05R,EXTRA=RGB01,START=2018-04-03,END=2018-04-26,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_115_05R" -v TRIAL=TR0115,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC05L,EXTRA=RGB02,START=2018-04-03,END=2018-04-26,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_115_05P" -v TRIAL=TR0115,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC05-Picam,START=2018-04-03,END=2018-04-26,RESOLUTION_HIRES="$RESOLUTION_HIGH_PICAM",RESOLUTION="$RESOLUTION_PICAM"
#
# qsub run_pipeline_single.pbs -a 0100 -N "proc_116_P" -v TRIAL=TR0116,START=2018-04-09,END=2018-04-27,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Picam,RESOLUTION_HIRES="$RESOLUTION_HIGH_PICAM",RESOLUTION="$RESOLUTION_PICAM"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_116_C01" -v TRIAL=TR0116,START=2018-04-09,END=2018-04-27,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam01,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"

# qsub run_pipeline_single.pbs -a 0100 -N "proc_117_P" -v TRIAL=TR0117,START=2018-04-27,END=2018-07-06,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Picam,RESOLUTION_HIRES="$RESOLUTION_HIGH_PICAM",RESOLUTION="$RESOLUTION_PICAM"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_117_C01" -v TRIAL=TR0117,START=2018-04-27,END=2018-07-06,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam01,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"

# qsub run_pipeline_single.pbs -a 0100 -N "proc_118_04L" -v TRIAL=TR0118,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC04R,START=2018-06-26,END=2018-08-17,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_118_04R" -v TRIAL=TR0118,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC04L,START=2018-06-26,END=2018-08-17,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_118_04P" -v TRIAL=TR0118,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC04-Picam,START=2018-06-26,END=2018-08-17,RESOLUTION_HIRES="$RESOLUTION_HIGH_PICAM",RESOLUTION="$RESOLUTION_PICAM"
#
# qsub run_pipeline_single.pbs -a 0100 -N "proc_119_35L" -v TRIAL=TR0119,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC35R,START=2018-06-26,END=2018-08-17,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_119_35R" -v TRIAL=TR0119,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC35L,START=2018-06-26,END=2018-08-17,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_119_35P" -v TRIAL=TR0119,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC35-Picam,START=2018-06-26,END=2018-08-17,RESOLUTION_HIRES="$RESOLUTION_HIGH_PICAM",RESOLUTION="$RESOLUTION_PICAM"
#
qsub run_pipeline_single.pbs -a 0100 -N "proc_120_P" -v TRIAL=TR0120,START=2018-07-16,END=2018-08-24,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Picam,RESOLUTION_HIRES="$RESOLUTION_HIGH_PICAM",RESOLUTION="$RESOLUTION_PICAM"
qsub run_pipeline_single.pbs -a 0100 -N "proc_120_C01" -v TRIAL=TR0120,START=2018-07-16,END=2018-08-24,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam01,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
#
# qsub run_pipeline_single.pbs -a 0100 -N "proc_121_03L" -v TRIAL=TR0121,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC03R,START=2018-06-26,END=2018-08-17,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_121_03R" -v TRIAL=TR0121,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC03L,START=2018-06-26,END=2018-08-17,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_121_03P" -v TRIAL=TR0121,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC03-Picam,START=2018-06-26,END=2018-08-17,RESOLUTION_HIRES="$RESOLUTION_HIGH_PICAM",RESOLUTION="$RESOLUTION_PICAM"

qsub run_pipeline_single.pbs -a 0100 -N "proc_122_P" -v TRIAL=TR0122,START=2018-09-03,END=2018-10-02,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Picam,RESOLUTION_HIRES="$RESOLUTION_HIGH_PICAM",RESOLUTION="$RESOLUTION_PICAM"
qsub run_pipeline_single.pbs -a 0100 -N "proc_122_C1" -v TRIAL=TR0122,START=2018-09-03,END=2018-10-02,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam01,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
qsub run_pipeline_single.pbs -a 0100 -N "proc_122_C2" -v TRIAL=TR0122,START=2018-09-03,END=2018-10-02,SOURCE=/home/801/gdd801/phenomics/camupload/picam/Eucalyptus02-Cam02,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"

# qsub run_pipeline_single.pbs -a 0100 -N "proc_123_37Ln" -v TRIAL=TR0123,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC37R-NIR-C02,START=2018-08-01,END=2018-10-19,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_123_37Rn" -v TRIAL=TR0123,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC37L-NIR-C01,START=2018-08-01,END=2018-10-19,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_123_37L" -v TRIAL=TR0123,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC37R,START=2018-08-01,END=2018-10-19,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_123_37R" -v TRIAL=TR0123,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC37L,START=2018-08-01,END=2018-10-19,RESOLUTION_HIRES="$RESOLUTION_HIGH_DSLR",RESOLUTION="$RESOLUTION_DSLR"
# qsub run_pipeline_single.pbs -a 0100 -N "proc_123_37P" -v TRIAL=TR0123,SOURCE=/home/801/gdd801/phenomics/camupload/picam/GC37-Picam,START=2018-08-01,END=2018-10-19,RESOLUTION_HIRES="$RESOLUTION_HIGH_PICAM",RESOLUTION="$RESOLUTION_PICAM"
