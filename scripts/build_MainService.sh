#!/bin/bash

# Dynamically set AppRoot to project root (two levels up from scripts/)
AppRoot=$(dirname "$(dirname "$(realpath "$0")")")

function moduleInit() {
  sh "$AppRoot/scripts/moduleInit.sh"
}

function make_all() {
  echo "Building the project"
  cd "$AppRoot/build" || exit
  make -f Makefile
  cd - || exit
}

function make_docker() {
  echo "Creating Docker Image"
  cd "$AppRoot/build" || exit
  make -f Makefile docker
  cd - || exit
}

# Argument handling
case "$1" in
  "make_all")
    moduleInit
    make_all
    ;;
  "make_docker")
    moduleInit
    make_docker
    ;;
  *)
    echo "Usage: $0 {make_all|make_docker}"
    exit 1
    ;;
esac