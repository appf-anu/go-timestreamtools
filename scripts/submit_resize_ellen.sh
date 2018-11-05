#!/bin/bash
SOURCE_START=/g/data1a/xe2/phenomics/timestreams_from_largedata/
qsub resize.pbs -N resize1 -v SOURCE=${SOURCE_START}/BVZ0059/outputs/BVZ0059-GC35L-C01~fullres-cor,NAME=BVZ0059-GC35L-C01~1920-cor
qsub resize.pbs -N resize2 -v SOURCE=${SOURCE_START}/BVZ0059/outputs/BVZ0059-GC35L-C01~fullres-seg,NAME=BVZ0059-GC35L-C01~1920-seg
qsub resize.pbs -N resize3 -v SOURCE=${SOURCE_START}/BVZ0059/outputs/BVZ0059-GC35R-C01~fullres-cor,NAME=BVZ0059-GC35R-C01~1920-cor
qsub resize.pbs -N resize4 -v SOURCE=${SOURCE_START}/BVZ0059/outputs/BVZ0059-GC35R-C01~fullres-seg,NAME=BVZ0059-GC35R-C01~1920-seg

qsub resize.pbs -N resize5 -v SOURCE=${SOURCE_START}/BVZ0060/outputs/BVZ0060-GC36R-C01~fullres-cor,NAME=BVZ0060-GC36R-C01~1920-cor
qsub resize.pbs -N resize6 -v SOURCE=${SOURCE_START}/BVZ0060/outputs/BVZ0060-GC36R-C01~fullres-seg,NAME=BVZ0060-GC36R-C01~1920-seg
qsub resize.pbs -N resize7 -v SOURCE=${SOURCE_START}/BVZ0060/outputs/BVZ0060-GC36L-C01~fullres-cor,NAME=BVZ0060-GC36L-C01~1920-cor
qsub resize.pbs -N resize8 -v SOURCE=${SOURCE_START}/BVZ0060/outputs/BVZ0060-GC36L-C01~fullres-seg,NAME=BVZ0060-GC36L-C01~1920-seg
