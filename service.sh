#!/bin/bash
set -e
export LIB=libduckdb.so 
export CGO_LDFLAGS="-L$(pwd)/lib" 
export LD_LIBRARY_PATH=/go/lib 
export CGO_CFLAGS="-I$(pwd)lib"
echo $CGO_LDFLAGS
case "$1" in
    develop)
        echo "Running Development Server"
        exec reflex -c ./reflex.conf
        ;;
    start)
        echo "Running Start"
        exec ./main
        ;;
    *)
        exec "$@"
esac