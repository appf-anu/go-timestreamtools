#!/bin/bash
# qsub archive.pbs -N "arch_Eu02-Picam" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/Eucalyptus02-Picam
# qsub archive.pbs -N "arch_Eu02-Cam01" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/Eucalyptus02-Cam01
# qsub archive.pbs -N "arch_GC02L" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC02L
# qsub archive.pbs -N "arch_GC02R" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC02R
# qsub archive.pbs -N "arch_GC02-Picam" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC02-Picam
# qsub archive.pbs -N "arch_GC03L" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC03L
# qsub archive.pbs -N "arch_GC03R" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC03R
# qsub archive.pbs -N "arch_GC03-Picam" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC03-Picam
# qsub archive.pbs -N "arch_GC04L" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC04L
# qsub archive.pbs -N "arch_GC04R" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC04R
# qsub archive.pbs -N "arch_GC04-Picam" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC04-Picam
# qsub archive.pbs -N "arch_GC05L" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC05L
# qsub archive.pbs -N "arch_GC05R" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC05R
# qsub archive.pbs -N "arch_GC05-Picam" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC05-Picam
# qsub archive.pbs -N "arch_GC35L" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC35L
# qsub archive.pbs -N "arch_GC35R" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC35R
# qsub archive.pbs -N "arch_GC35-Picam" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC35-Picam
# qsub archive.pbs -N "arch_GC36L" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC36L
# qsub archive.pbs -N "arch_GC36R" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC36R
# qsub archive.pbs -N "arch_GC36-Picam" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC36-Picam
# qsub archive.pbs -N "arch_GC37L" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC37L
# qsub archive.pbs -N "arch_GC37R" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC37R
# qsub archive.pbs -N "arch_GC37L-NIR-C01" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC37L-NIR-C01
# qsub archive.pbs -N "arch_GC37R-NIR-C02" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC37R-NIR-C02
# qsub archive.pbs -N "arch_GC37-Picam" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/GC37-Picam
qsub archive.pbs -l walltime=1:00:00 -N "arch_BVZ-HousePicam" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/BVZ-HousePicam
qsub archive.pbs -l walltime=00:20:00 -N "arch_MakerspacePicam01" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/MakerspacePicam01
qsub archive.pbs -N "arch_TimHouse-SpringCam01" -v SOURCE=/g/data1a/xe2/phenomics/camupload/picam/TimHouse-SpringCam01
