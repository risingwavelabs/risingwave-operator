#!/usr/bin/env bash

set -euo pipefail

# Create a directory to store the environment variables
mkdir -p ci/run && env | grep "^BUILDKITE\|^CI\|^AWS\|^RW" > ci/run/env.list
echo CI_ENV=1 >> ci/run/env.list

# Get environment variables from the command line
env_args=""
while getopts ":e:" opt; do
	case ${opt} in
	e)
		[[ -v ${OPTARG} ]] && env_args+="-e ${OPTARG} "
		;;
	\?)
		echo "Invalid option: $OPTARG" 1>&2
		exit 1
		;;
	:)
		echo "Invalid option: $OPTARG requires an argument" 1>&2
		exit 1
		;;
	esac
done

# Get the arguments after the options
shift $((OPTIND - 1))

# shellcheck disable=SC2086
docker run --rm --userns=host --privileged --network host \
	${env_args} --env-file ci/run/env.list \
	-v "$(pwd):/workspace" -w /workspace \
	-v /var/run/docker.sock:/var/run/docker.sock \
	--entrypoint bash \
	"${CI_BASE_IMAGE}" "$@"
