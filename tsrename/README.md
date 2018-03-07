# tsrename
image renaming program written in Go

Is intended to be used in conjunction with tsarchive, tsresize
usage of ./tsrename:

	copy with <name> prefix:
		 ./tsrename -source <source> -name=<name>
	copy with <name> prefix:
		 ./tsrename -source <source> -name=<name>

flags:
	-del: removes the source files
	-name: renames the prefix fo the target files
	-exif: uses exif data to rename rather than file timestamp
	-output: set the <destination> directory (default=.)
	-source: set the <source> directory (optional, default=stdin)

reads filepaths from stdin
writes paths to resulting files to stdout
will ignore any line from stdin that isnt a filepath (and only a filepath)
