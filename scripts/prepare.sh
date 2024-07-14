#!/bin/sh

set -x 

FAMILY=""
ARCHITECTURE=""


case "$(uname -s)" in
    CYGWIN*) FAMILY="cygwin" ;;
    Linux*) 
        . /etc/os-release
        case ${ID} in
            ubuntu) FAMILY="linux" ;;
            alpine) FAMILY="alpine" ;;
            *) FAMILY="linux" ;;
        esac
        ;;
    *) FAMILY="unknown" ;;
esac

case "$(uname -m)" in 
  amd64|x86_64)   ARCHITECTURE="amd64" ;; 
  *) ARCHITECTURE="x86" ;; 
esac



if [ -z "${BUILD_ID}" ]; then
    BUILD_ID="(none)"
    VERSION="0.0.1-SNAPSHOT"
    REPOSITORY=snapshots
    REPOSITORYID=snapshots
else
    VERSION="0.0.1.$((${BUILD_ID}))"
    REPOSITORY=releases
    REPOSITORYID=releases
fi




BASEDIR=$(dirname "$0")
SCRIPT_DIR=$(cd $BASEDIR && pwd)
PROJECT_DIR=$(dirname $SCRIPT_DIR)
SOURCE_DIR=${PROJECT_DIR}/src
BUILD_DIR=${PROJECT_DIR}/build






PROJECT=diaries-responder-go
GROUPID=com.rsmaxwell.diaries
ARTIFACTID=${PROJECT}_${FAMILY}_${ARCHITECTURE}
PACKAGING=zip
ZIPFILE=${ARTIFACTID}_${VERSION}.${PACKAGING}



TIMESTAMP="$(date '+%Y-%m-%d %H:%M:%S')"
GIT_COMMIT="${GIT_COMMIT:-(none)}"
GIT_BRANCH="${GIT_BRANCH:-(none)}"
GIT_URL="${GIT_URL:-(none)}"

export PROJECT
export REPOSITORY
export REPOSITORYID
export BUILD_ID
export VERSION
export TIMESTAMP
export GIT_COMMIT
export GIT_BRANCH
export GIT_URL
export FAMILY
export ARCHITECTURE


cd ${SOURCE_DIR}

tags='$FAMILY,$ARCHITECTURE,$PROJECT,$REPOSITORY,$REPOSITORYID,$VERSION,$BUILD_ID,$TIMESTAMP,$GIT_COMMIT,$GIT_BRANCH,$GIT_URL'

find . -name BuildInfo.go | while read filename; do
    echo "Writing ${filename}"
    file=${SOURCE_DIR}/${filename}
    dir=${directory ${file}}
    mkdir -p ${dir}
    envsubst "${tags}" < ${filename} > ${file}
done




mkdir -p ${BUILD_DIR}
cd ${BUILD_DIR}

cat > buildinfo <<EOL
BUILD_ID="${BUILD_ID}"
VERSION="${VERSION}"
REPOSITORY="${REPOSITORY}"
REPOSITORYID="${REPOSITORYID}"
REPOSITORY_URL="${REPOSITORY_URL}"
FAMILY="${FAMILY}"
ARCHITECTURE="${ARCHITECTURE}"
PROJECT="${PROJECT}"
GROUPID="${GROUPID}"
ARTIFACTID="${ARTIFACTID}"
PACKAGING="${PACKAGING}"
ZIPFILE="${ZIPFILE}"
TIMESTAMP="${TIMESTAMP}"
GIT_COMMIT="${GIT_COMMIT}"
GIT_BRANCH="${GIT_BRANCH}"
GIT_URL="${GIT_URL}"
EOL

pwd
ls -al 
cat buildinfo
