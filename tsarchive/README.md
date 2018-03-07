# tsarchive
archives images by week

usage of ./tsarchive:

	archive files from directory:
		 ./tsarchive -source <source> -output <output>

flags:

	-output: set the <destination> directory (default=%s)
	-source: set the <source> directory (optional, default=stdin)
	-name: set the name prefix of the output tarfile <name>2006-01-02.tar (default=guess)

reads filepaths from stdin
writes paths to resulting files to stdout
will ignore any line from stdin that isnt a filepath (and only a filepath)
