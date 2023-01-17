#!/bin/bash

get_args() {
	_arg=$1
	shift
	echo "$@" | grep -Eo "\-\-${_arg}=[^ ]+" | cut -d= -f2 | tail -n 1
	unset _arg
}

configPath="./settings/sources.json"

[ "$GRAVITY_ADAPTER_STAN_SOURCE_CONFIG" != "" ] && {
	configPath=$GRAVITY_ADAPTER_STAN_SOURCE_CONFIG
} 

sources=$(get_args sources "$@")
[ "$sources" != "" ] && {
	echo $sources > $configPath
}

[ "$sources" != "" ] || {
	echo $GRAVITY_ADAPTER_STAN_SOURCE_SETTINGS > $configPath
}

exec /gravity-adapter-stan
