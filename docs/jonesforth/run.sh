#!/bin/sh
docker build -t jonesforth .
docker run --cap-add=SYS_RAWIO -ti --rm jonesforth
