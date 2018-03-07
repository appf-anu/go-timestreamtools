# tsrename
image renaming program written in Go

Is intended to be used in conjunction with tsarchive, tsresize


usage of ./tsselect:

	filter from 11 June 1996 until now with source:
		 ./tsselect -source <source> -start 1996-06-11
	filter from 11 June 1996 to 10 December 1996 from stdin:
		 ./tsselect -start 1996-06-11 -end 1996-12-10

flags:

	-start: the start datetime (default=1970-01-01 00:00)
	-end: the end datetime (default=now)
	-exif: uses exif data to get time instead of the file timestamp
	-source: set the <source> directory (optional, default=stdin)

dates are assumed to be DMY or YMD not MDY

reads filepaths from stdin
writes paths to resulting files to stdout
will ignore any line from stdin that isnt a filepath (and only a filepath)
