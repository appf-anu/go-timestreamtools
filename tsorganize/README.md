# tsrename
image renaming program written in Go

Is intended to be used in conjunction with tsarchive, tsresize

usage of ./tsorganize:

	copy into structure:
		 ./tsorganize -source <source>
	copy into structure at <destination>:
		 ./tsorganize -source <source> -output=<destination>
	rename (move) into structure:
		 ./tsorganize -source <source> -del

flags:

	-del: removes the source files
	-dirstruct: directory structure to pass to golangs time.Format
	-exif: uses exif data to rename rather than file timestamp
	-output: set the <destination> directory (default=.)
	-source: set the <source> directory (optional, default=stdin)

reads filepaths from stdin
writes paths to resulting files to stdout
will ignore any line from stdin that isnt a filepath (and only a filepath)
