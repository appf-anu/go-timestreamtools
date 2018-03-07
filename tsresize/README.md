# tsresize
go program to resize images
usage of ./tsresize:

flags:

	-res: output image resolution
	-output: <destination> directory (default=.)
	-type: output image type (default=jpeg)

		available image types:

		jpeg, png
		tiff: tiff with Deflate compression (alias for tiff-deflate)
		tiff-none: tiff with no compression

reads filepaths from stdin
writes paths to resulting files to stdout
will ignore any line from stdin that isnt a filepath (and only a filepath)
