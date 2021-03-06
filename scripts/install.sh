#!/bin/bash
#
# Tilt installer
#
# Usage:
#   curl -fsSL https://raw.githubusercontent.com/windmilleng/tilt/master/scripts/install.sh | bash

# When releasing Tilt, the releaser should update this version number
# AFTER they upload new binaries.
VERSION="0.12.2"
BREW=$(which brew)

set -e

function install_tilt() {
  if [[ "$OSTYPE" == "linux-gnu" ]]; then
      set -x
      curl -fsSL https://github.com/windmilleng/tilt/releases/download/v$VERSION/tilt.$VERSION.linux.x86_64.tar.gz | tar -xzv tilt
      sudo mv tilt /usr/local/bin/tilt
  elif [[ "$OSTYPE" == "darwin"* ]]; then
      if [[ "$BREW" != "" ]]; then
          set -x
          brew tap windmilleng/tap
          brew install windmilleng/tap/tilt
      else
          set -x
          curl -fsSL https://github.com/windmilleng/tilt/releases/download/v$VERSION/tilt.$VERSION.mac.x86_64.tar.gz | tar -xzv tilt
          sudo mv tilt /usr/local/bin/tilt
      fi
  else
      set +x
      echo "The Tilt installer does not work for your platform: $OS"
      echo "Please file an issue at https://github.com/windmilleng/tilt/issues/new"
      exit 1
  fi

  set +x
}

function version_check() {
  VERSION="$(tilt version 2>&1 || true)"
  RUBY_TILT_PATTERN="template engine not found"
  TILT_DEV_PATTERN='^v[0-9]+\.[0-9]+\.[0-9]+(-dev)?, built [0-9]+-[0-9]+-[0-9]+$'
  if [[ $VERSION =~ $RUBY_TILT_PATTERN ]]; then
    echo "Tilt installed!"
    echo
    echo "Note: the ruby templating program named 'tilt' (at $(command -v tilt)) appears before tilt.dev's tilt in your \$PATH."
    echo "You'll need to adjust your \$PATH, uninstall the other tilt, or use an absolute path to run tilt.dev's tilt."
    exit 1
  elif ! [[ $VERSION =~ $TILT_DEV_PATTERN ]]; then
    echo "Tilt installed!"
    echo
    echo "Note: it looks like it is not the first program named 'tilt' in your path. \`tilt version\` (running from $(command -v tilt)) did not return a a tilt.dev version string."
    echo "It output this instead:"
    echo
    echo "$VERSION"
    echo
    echo "Perhaps you have a different program named tilt in your \$PATH?"
    exit 1
  else
    echo "Tilt installed! Run \`tilt up\` to start."
  fi
}
  
# so that we can skip installation in CI and just test the version check
if [[ -z $NO_INSTALL ]]; then
  install_tilt
fi

version_check

