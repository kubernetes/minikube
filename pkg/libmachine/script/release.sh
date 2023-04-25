#!/bin/bash

GITHUB_USER=docker
GITHUB_REPO=machine
PROJECT_URL="git@github.com:${GITHUB_USER}/${GITHUB_REPO}"

function usage {
  echo "Usage: "
  echo "   GITHUB_TOKEN=XXXXX ${0} [X.Y.Z]"
}

function display {
  echo "🐳  $1"
  echo
}

function checkError {
  if [[ "$?" -ne 0 ]]; then
    echo "😡   $1"
    exit 1
  fi
}

function createMachine {
  docker-machine rm -f release 2> /dev/null
  docker-machine create -d virtualbox --virtualbox-cpu-count=2 --virtualbox-memory=2048 release
}

if [[ -z "${GITHUB_TOKEN}" ]]; then
  echo "GITHUB_TOKEN missing"
  usage
  exit 1
fi

VERSION=$1

if [[ -z "${VERSION}" ]]; then
  echo "Missing version argument"
  usage
  exit 1
fi

if [[ ! "${VERSION}" =~ ^[0-9]\.[0-9]+(\.[0-9])+(-rc[1-9][0-9]*)?$ ]]; then
  echo "Invalid version. It should look like 0.5.1, 0.6 or 0.5.1-rc2"
  exit 1
fi

command -v git > /dev/null 2>&1
checkError "You obviously need git, please consider installing it..."

command -v github-release > /dev/null 2>&1
checkError "github-release is not installed, go get -u github.com/aktau/github-release or check https://github.com/aktau/github-release, aborting."

command -v openssl > /dev/null 2>&1
checkError "You need openssl to generate binaries signature, brew install it, aborting."

command -v docker-machine > /dev/null 2>&1
checkError "You must have a docker-machine in your path"

GITHUB_VERSION="v${VERSION}"
RELEASE_DIR="$(dirname "$(git rev-parse --show-toplevel)")/release-${VERSION}"
GITHUB_RELEASE_FILE="github-release-${VERSION}.md"

LAST_RELEASE_VERSION=$(git describe --tags $(git rev-list --tags --max-count=1))
checkError "Unable to find current version tag"

display "Starting release from ${LAST_RELEASE_VERSION} to ${GITHUB_VERSION} on ${PROJECT_URL} with token ${GITHUB_TOKEN}"
if [[ "${GITHUB_USER}" == "docker" ]]; then
    display "THIS IS A REAL RELEASE, on OFFICIAL DOCKER REPO"
fi
while true; do
    read -p "🐳  Do you want to proceed with this release? (y/n) > " yn
    echo ""
    case $yn in
        [Yy]* ) break;;
        [Nn]* ) exit;;
        * ) echo "😡   Please answer yes or no.";;
    esac
done

display "Checking machine 'release' status"
MACHINE_STATUS=$(docker-machine status release)
if [[ "$?" -ne 0 ]]; then
  display "Machine 'release' does not exist, creating it"
  createMachine
else
  if [[ "${MACHINE_STATUS}" != "Running" ]]; then
    display "Machine 'release' is not running, trying to start it"
    docker-machine start release
    docker-machine env release
    if [[ "$?" -ne 0 ]]; then
      display "Machine 'release' could not be started, removing and creating a fresh new one"
      createMachine
    fi
    display "Loosing 5 seconds to the VirtualBox gods"
    sleep 5
  fi
fi

eval "$(docker-machine env release)"
checkError "Machine 'release' is in a weird state, aborting"

if [[ -d "${RELEASE_DIR}" ]]; then
  display "Cleaning up ${RELEASE_DIR}"
  rm -rdf "${RELEASE_DIR}"
  checkError "Can't clean up ${RELEASE_DIR}. You should do it manually and retry"
fi

display "Cloning into ${RELEASE_DIR} from ${PROJECT_URL}"

mkdir -p "${RELEASE_DIR}"
checkError "Can't create ${RELEASE_DIR}, aborting"
git clone -q "${PROJECT_URL}" "${RELEASE_DIR}"
checkError "Can't clone into ${RELEASE_DIR}, aborting"

cd "${RELEASE_DIR}"

display "Bump version number to ${VERSION}"

# Why 'sed' and then 'mv' instead of 'sed -i'?  BSD / GNU sed compatibility.
# Macs have BSD sed by default, Linux has GNU sed.  See
# http://unix.stackexchange.com/questions/92895/how-to-achieve-portability-with-sed-i-in-place-editing
sed -e "s/Version = \".*\"$/Version = \"${VERSION}\"/g" version/version.go >version/version.go.new
checkError "Unable to change version in version/version.go"
mv -- version/version.go.new version/version.go

git add version/version.go
git commit -q -m"Bump version to ${VERSION}" -s
checkError "Can't git commit the version upgrade, aborting"

display "Building in-container style"
USE_CONTAINER=true make clean build validate build-x
checkError "Build error, aborting"

display "Generating github release"
cp -f script/release/github-release-template.md "${GITHUB_RELEASE_FILE}"
checkError "Can't find github release template"
CONTRIBUTORS=$(git log "${LAST_RELEASE_VERSION}".. --format="%aN" --reverse | sort | uniq | awk '{printf "- %s\n", $0 }')
CHANGELOG=$(git log "${LAST_RELEASE_VERSION}".. --oneline | grep -v 'Merge pull request')

CHECKSUM=""
rm -f sha256sum.txt md5sum.txt
for file in $(ls bin/docker-machine-*); do
  SHA256=$(openssl dgst -sha256 < "${file}")
  MD5=$(openssl dgst -md5 < "${file}")
  LINE=$(printf "\n * **%s**\n  * sha256 \`%s\`\n  * md5 \`%s\`\n\n" "$(basename ${file})" "${SHA256}" "${MD5}")
  CHECKSUM="${CHECKSUM}${LINE}"
  echo "${SHA256}  ${file:4}" >> sha256sum.txt
  echo "${MD5}  ${file:4}" >> md5sum.txt
done

TEMPLATE=$(cat "${GITHUB_RELEASE_FILE}")
echo "${TEMPLATE//\{\{VERSION\}\}/$GITHUB_VERSION}" > "${GITHUB_RELEASE_FILE}"
checkError "Couldn't replace [ ${GITHUB_VERSION} ]"

TEMPLATE=$(cat "${GITHUB_RELEASE_FILE}")
echo "${TEMPLATE//\{\{CHANGELOG\}\}/$CHANGELOG}" > "${GITHUB_RELEASE_FILE}"
checkError "Couldn't replace [ ${CHANGELOG} ]"

TEMPLATE=$(cat "${GITHUB_RELEASE_FILE}")
echo "${TEMPLATE//\{\{CONTRIBUTORS\}\}/$CONTRIBUTORS}" > "${GITHUB_RELEASE_FILE}"
checkError "Couldn't replace [ ${CONTRIBUTORS} ]"

TEMPLATE=$(cat "${GITHUB_RELEASE_FILE}")
echo "${TEMPLATE//\{\{CHECKSUM\}\}/$CHECKSUM}" > "${GITHUB_RELEASE_FILE}"
checkError "Couldn't replace [ ${CHECKSUM} ]"

RELEASE_DOCUMENTATION="$(cat ${GITHUB_RELEASE_FILE})"

display "Tagging and pushing tags"
git remote | grep -q remote.prod.url
if [[ "$?" -ne 0 ]]; then
  display "Adding 'remote.prod.url' remote git url"
  git remote add remote.prod.url "${PROJECT_URL}"
fi

display "Checking if remote tag ${GITHUB_VERSION} already exists"
git ls-remote --tags 2> /dev/null | grep -q "${GITHUB_VERSION}" # returns 0 if found, 1 if not
if [[ "$?" -ne 1 ]]; then
  display "Deleting previous tag ${GITHUB_VERSION}"
  git tag -d "${GITHUB_VERSION}" &> /dev/null
  git push -q origin :refs/tags/"${GITHUB_VERSION}"
else
  echo "Tag ${GITHUB_VERSION} does not exist... yet"
fi

display "Tagging release on github"
git tag "${GITHUB_VERSION}"
git push -q remote.prod.url "${GITHUB_VERSION}"
checkError "Could not push to remote url"

display "Checking if release already exists"
github-release info \
    --security-token  "${GITHUB_TOKEN}" \
    --user "${GITHUB_USER}" \
    --repo "${GITHUB_REPO}" \
    --tag "${GITHUB_VERSION}" > /dev/null 2>&1

if [[ "$?" -ne 1 ]]; then
  display "Release already exists, cleaning it up"
  github-release delete \
      --security-token  "${GITHUB_TOKEN}" \
      --user "${GITHUB_USER}" \
      --repo "${GITHUB_REPO}" \
      --tag "${GITHUB_VERSION}"
  checkError "Could not delete release, aborting"
fi

display "Creating release on github"
github-release release \
    --security-token  "${GITHUB_TOKEN}" \
    --user "${GITHUB_USER}" \
    --repo "${GITHUB_REPO}" \
    --tag "${GITHUB_VERSION}" \
    --name "${GITHUB_VERSION}" \
    --description "${RELEASE_DOCUMENTATION}" \
    --pre-release
checkError "Could not create release, aborting"

display "Uploading binaries"
for file in $(ls bin/docker-machine-*); do
  display "Uploading ${file}..."
  github-release upload \
      --security-token  "${GITHUB_TOKEN}" \
      --user "${GITHUB_USER}" \
      --repo "${GITHUB_REPO}" \
      --tag "${GITHUB_VERSION}" \
      --name "$(basename "${file}")" \
      --file "${file}"
  if [[ "$?" -ne 0 ]]; then
    display "Could not upload ${file}, continuing with others"
  fi
done

display "Uploading sha256sum.txt and md5sum.txt"
for file in sha256sum.txt md5sum.txt; do
  display "Uploading ${file}..."
  github-release upload \
      --security-token  "${GITHUB_TOKEN}" \
      --user "${GITHUB_USER}" \
      --repo "${GITHUB_REPO}" \
      --tag "${GITHUB_VERSION}" \
      --name "$(basename "${file}")" \
      --file "${file}"
  if [[ "$?" -ne 0 ]]; then
    display "Could not upload ${file}, continuing with others"
  fi
done

git remote rm remote.prod.url

rm ${GITHUB_RELEASE_FILE}

echo "There is a couple of tasks your still need to do manually:"
echo "  1. Open the release notes created for you on github https://github.com/${GITHUB_USER}/${GITHUB_REPO}/releases/tag/${GITHUB_VERSION}, you'll have a chance to enhance commit details a bit"
echo "  2. Once you're happy with your release notes on github, copy the list of changes to the CHANGELOG.md"
echo "  3. Update the documentation branch"
echo "  4. Test the binaries linked from the github release page"
echo "  5. Change version/version.go to the next dev version"
echo "  6. Party !!"
