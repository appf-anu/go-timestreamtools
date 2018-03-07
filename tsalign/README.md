# tsalign
aligns image files to a specified or assumed interval.

usage of ./tsalign:

	align images in place:
		./tsalign -source <source> -output <source>
	 copy aligned to <destination>:
		./tsalign -source <source> -output=<destination>

flags:

	-name: renames the prefix fo the target files
	-exif: uses exif data to rename rather than file timestamp
	-output: set the <destination> directory (default=<cwd>)	-source: set the <source> directory (optional, default=stdin)
	-interval: set the interval to align to (optional, default=5m)

will only align down, if an image is at 10:03 (5m interval) it will align to 10:00
chronologically earlier images will be kept
ie. at 5m interval, an image at 10:03 will overwrite an image at 10:02

reads filepaths from stdin
writes paths to resulting files to stdout
will ignore any line from stdin that isnt a filepath (and only a filepath)
