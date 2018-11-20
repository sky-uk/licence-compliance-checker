#!/bin/bash -e

cleanup () {
    set +e
    docker rm -f ${containerName}
    rm -f ${cidfile}
}
trap "{ CODE=$?; cleanup ; exit $CODE; }" EXIT

outputDir=build
containerName=licence-compliance-checker-test
cidfile=${outputDir}/cid
buildImageName="local/licence-compliance-checker:build"

echo "Building licence-compliance-checker"
docker build -t ${buildImageName} . -f Dockerfile.build

mkdir -p ${outputDir}
echo "Running tests within $buildImageName"

set +e
docker run --cidfile ${cidfile} --name ${containerName} ${buildImageName} make check
testRunExitCode=$?
set -e

docker cp ${containerName}:/go/src/github.com/sky-uk/licence-compliance-checker/build/junit-reports ${outputDir}
exit ${testRunExitCode}
