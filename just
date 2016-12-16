#! /bin/bash

set -eu
cd `dirname $0`

NAME="bitvishttp"
VERSION="$(git describe --always --dirty) ($(date --date="@$(git show -s --format='%ct' HEAD)" '+%Y-%m-%d'))"
WORKSPACE="$PWD"
BIN="$WORKSPACE/bin"
LIB="$WORKSPACE/lib"
GOPATH="$LIB/gopath"

mkdir -p "$BIN"
mkdir -p "$LIB"

usage() {
cat << EOF
SYNOPSIS

    This is a very handy tool to manage this Go project.

USAGE
      $ $0 install    # Install dependencies
      $ $0 build      # Build a binary
      $ $0 run        # Run the currently built binary
EOF
}

case ${1:-} in
	"install")
		echo "*** Installing dependencies ***"
		cat "$WORKSPACE/src/Godeps" | \
		while read dep; do
			pkg=`echo $dep | cut -f1 -d' '`
			rev=`echo $dep | cut -f2 -d' '`
			echo "  Installing $pkg"
			GOPATH="$GOPATH" go get "$pkg"
			pushd "$GOPATH/src/$pkg" > /dev/null; git checkout $rev --quiet; popd > /dev/null
		done
	;;

	"build")
		echo "*** Building Project ***"
		cd "$WORKSPACE/src"
		GOPATH="$GOPATH" go build -o "$WORKSPACE/bin/$NAME"
		cd "$WORKSPACE"
	;;

	"run")
		"$BIN/$NAME"
	;;

	*)
		usage
	;;
esac
