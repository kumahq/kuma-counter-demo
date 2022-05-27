#!/usr/bin/env bash

set -e

KUMA_DOCKER_REPO="${KUMA_DOCKER_REPO:-docker.io}"
KUMA_DOCKER_REPO_ORG="${KUMA_DOCKER_REPO_ORG:-${KUMA_DOCKER_REPO}/kumahq}"
BUILD_ARCH="${BUILD_ARCH:-amd64 arm64}"
VERSION="${VERSION:-latest}"
DRY_RUN="${DRY_RUN:-true}"

function build_and_push() {
  for arch in ${BUILD_ARCH}; do
    echo "Building kuma-demo..."
    docker build --pull --build-arg ARCH="${arch}" -t "${KUMA_DOCKER_REPO_ORG}/kuma-demo:latest-${arch}" -f release/Dockerfile .
    if [ $VERSION != "latest" ]; then 
      docker tag "${KUMA_DOCKER_REPO_ORG}/kuma-demo:${VERSION}-${arch}" "${KUMA_DOCKER_REPO_ORG}/kuma-demo:latest-${arch}"
    fi
    if [ $DRY_RUN != "true" ]; then
      echo "Pushing kuma-demo:$VERSION-$arch ..."
      docker push "${KUMA_DOCKER_REPO_ORG}/kuma-demo:${VERSION}-${arch}"
      echo "... done!"
    fi  
    echo "... done!"   
  done
  if [ $DRY_RUN != "true" ]; then
      images=()
      for arch in ${BUILD_ARCH}; do
        images+=("--amend ${KUMA_DOCKER_REPO_ORG}/kuma-demo:${VERSION}-${arch}")
      done
      command="docker manifest create ${KUMA_DOCKER_REPO_ORG}/kuma-demo:${VERSION} ${images[*]}"
      echo "Creating manifest for ${KUMA_DOCKER_REPO_ORG}/kuma-demo:${VERSION}..."
      eval "$command"
      echo "Pushing manifest ${KUMA_DOCKER_REPO_ORG}/kuma-demo:${VERSION} ..."
      docker manifest push "${KUMA_DOCKER_REPO_ORG}/kuma-demo:${VERSION}"
      echo ".. done!"
  fi
}

if [ $DRY_RUN != "true" ]; then
  [ -z "$DOCKER_USERNAME" ] && echo "\$DOCKER_USERNAME required"
  [ -z "$DOCKER_API_KEY" ] && echo "\$DOCKER_API_KEY required"
fi
build_and_push
