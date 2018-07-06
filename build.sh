#!/bin/bash

case $1 in
  archive)
    mkdir -p bin/
    zip -r bin/connect4.zip connect4/*.go
    ;;

  test)
    go test github.com/hjfreyer/aigames/connect4/testing
    ;;

  *)
    echo "Unknown command '$1'"
esac
