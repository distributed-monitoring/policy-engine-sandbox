#!/usr/bin/env bash
set -e

if [ "$(uname)" == "Darwin" ]; then
	export GOOS=linux
fi

ORG_PATH="github.com/distributed-monitoring"
export REPO_PATH="${ORG_PATH}/policy-engine-sandbox"

if [ ! -h gopath/src/${REPO_PATH} ]; then
	mkdir -p gopath/src/${ORG_PATH}
	ln -s ../../../.. gopath/src/${REPO_PATH} || exit 255
fi

export GOPATH=${PWD}/gopath

mkdir -p "${PWD}/bin"

echo "Building cmds"
CMDS="cmd/*"
for d in $CMDS; do
	if [ -d "$d" ]; then
		cmd="$(basename "$d")"
		echo "  $cmd"
		# use go install so we don't duplicate work
		if [ -n "$FASTBUILD" ]
		then
			GOBIN=${PWD}/bin go install -pkgdir $GOPATH/pkg "$@" $REPO_PATH/$d
		else
			go build -o "${PWD}/bin/$cmd" -pkgdir "$GOPATH/pkg" "$@" "$REPO_PATH/$d"
		fi
	fi
done
