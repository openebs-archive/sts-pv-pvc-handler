# This script builds the application from source for multiple platforms.
set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/../" && pwd )"

# Change into that directory
cd "$DIR"

## Populate the version based on release tag
## If release tag is set then assign it as VERSION and
## if release tag is empty then mark version as ci
if [ -n "$RELEASE_TAG" ]; then
    # Trim the `v` from the RELEASE_TAG if it exists
    # Example: v1.10.0 maps to 1.10.0
    # Example: 1.10.0 maps to 1.10.0
    # Example: v1.10.0-custom maps to 1.10.0-custom
    VERSION="${RELEASE_TAG#v}"
else
    VERSION="ci"
fi

echo "Building for ${VERSION} VERSION"

# Determine the arch/os combos we're building for
UNAME=$(uname)
if [ "$UNAME" != "Linux" -a "$UNAME" != "Darwin" ] ; then
    echo "Sorry, this OS is not supported yet."
    exit 1
fi

if [ "$UNAME" = "Darwin" ] ; then
  XC_OS="darwin"
elif [ "$UNAME" = "Linux" ] ; then
  XC_OS="linux"
fi


if [ -z "${PNAME}" ];
then
    echo "Project name not defined"
    exit 1
fi

if [ -z "${CTLNAME}" ];
then
    echo "CTLNAME not defined"
    exit 1
fi

# Delete the old dir
echo "==> Removing old directory..."
rm -rf bin/${PNAME}/*
mkdir -p bin/${PNAME}/

# If its dev mode, only build for ourself
if [[ "${DEV}" ]]; then
    XC_OS=$(go env GOOS)
    XC_ARCH=$(go env GOARCH)
fi

if [ -z "${XC_ARCH}" ];
then
    XC_ARCH=$(go env GOARCH)
fi

# Build!
echo "==> Building ${CTLNAME} using $(go version)... "

GOOS="${XC_OS}"
GOARCH="${XC_ARCH}"

output_name="bin/${PNAME}/${GOOS}_${GOARCH}/${CTLNAME}"

if [ $GOOS = "windows" ]; then
    output_name+='.exe'
fi

env GOOS=$GOOS GOARCH=$GOARCH go build ${BUILD_TAG} -o $output_name .

echo ""

# Done!
echo
echo "==> Results:"
ls -hl bin/${PNAME}/
