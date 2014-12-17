#!/usr/bin/env bash
# this requires the docker deb package to be built first. You can build it with
# `make all`
#
# To upload the releases, the build environment needs the following ENV
# - UPLOAD_DEB=true
# - DEBIAN_REPO=apt.nitrous.io(default)
# - APT_GPG_PASSPHRASE
# - APT_GPG_PUBLIC
# - APT_GPG_SECRET
set -e
die() {
  echo $* && exit 1
}
# by default, we only allow release via docker
IN_DOCKER=${1="true"}

# check if we should upload the artifacts to private debian repo or s3
UPLOAD_DEB=${UPLOAD_DEB:="false"}

# required environment variables for uploading the artifacts
if [[ "$UPLOAD_DEB" == "true" ]]; then
  APT_GPG_PASSPHRASE=${APT_GPG_PASSPHRASE:?"APT_GPG_PASSPHRASE is required"}
  APT_GPG_PUBLIC=${APT_GPG_PUBLIC:?"APT_GPG_PUBLIC is required"}
  APT_GPG_SECRET=${APT_GPG_SECRET:?"APT_GPG_SECRET is required"}
  ACCESS_KEY=${ACCESS_KEY:?"ACCESS_KEY is required"}
  SECRET_KEY=${SECRET_KEY:?"SECRET_KEY is required"}
  DEBIAN_REPO=${DEBIAN_REPO:?"DEBIAN_REPO is required"}
fi

bundle_docker() {
  VERSION=$(cat VERSION)
  for DEB in bundles/$VERSION/ubuntu/*.deb; do
    echo "Adding $DEB to local repository"
    deb-upload $DEBIAN_REPO $DEB public 6B19EAE0
  done
}

if [ "$IN_DOCKER" == "true" ]; then
  if [[ "$UPLOAD_DEB" == "true" ]]; then
    docker run \
      -v $APT_GPG_PASSPHRASE:/get.nitrous.io/.gpg/nitrous.passphrase.gpg \
      -e APT_GPG_PASSPHRASE=/get.nitrous.io/.gpg/nitrous.passphrase.gpg \
      -v $APT_GPG_PUBLIC:/get.nitrous.io/.gpg/nitrous.pub.gpg \
      -e APT_GPG_PUBLIC=/get.nitrous.io/.gpg/nitrous.pub.gpg \
      -v $APT_GPG_SECRET:/get.nitrous.io/.gpg/nitrous.sec.gpg \
      -e APT_GPG_SECRET=/get.nitrous.io/.gpg/nitrous.sec.gpg \
      -e UPLOAD_DEB=true \
      -e ACCESS_KEY=$ACCESS_KEY \
      -e SECRET_KEY=$SECRET_KEY \
      -e DEBIAN_REPO=$DEBIAN_REPO \
      -v $(pwd)/bundles:/root/bundles \
      -v $(pwd)/hack/nitrous-release.sh:/root/release \
      -v $(pwd)/VERSION:/root/VERSION \
      -w /root \
      --rm \
      nitrousio/packager /bin/bash -c "/root/release false"
  else
    echo "files to be uploaded:"
    VERSION=$(cat VERSION)
    for DEB in ./bundles/$VERSION/ubuntu/*.deb; do
      echo " - $DEB"
    done
  fi
else
  (bundle_docker)
fi
