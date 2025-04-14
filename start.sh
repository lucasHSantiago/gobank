#!/bin/sh

#to exit automatic if the return is not 0
set -e

echo "start the app"

#take all the parameters passed to the script and run it
exec "$@"
