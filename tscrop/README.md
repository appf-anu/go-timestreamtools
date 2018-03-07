# tscrop
crop sections out of an image

usage of ./tscrop:

centered crop to 1920x1080
	./tscrop -center -c1 1920,1080
cut out 120,10 to 400,60
	./tscrop -c1 120,10 c2 400,60
centered crop to 1920x1080 and output to <destination>
	./tscrop -center -c1 1920,1080 -output <destination>

flags:

	-center: center the crop, specify width,height with c1
	-c1: corner 1 (in pixels, comma separated)
	-c2: corner 2 (in pixels, comma separated, ignored if center is specified)
	-grid: split the area into this many equal crops (default=1,1)
	-type: set the output image type (default=jpeg)
		available image types:

		jpeg, png
		tiff: tiff with Deflate compression (alias for tiff-deflate)
		tiff-none: tiff with no compression
	-output: set the <destination> directory (default=<cwd>/<crop>)

reads filepaths from stdin
writes paths to resulting files to stdout
will ignore any line from stdin that isnt a filepath (and only a filepath)
