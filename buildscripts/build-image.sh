set -e

# Get the parent directory of where this script is.
SOURCE="${BASH_SOURCE[0]}"
while [ -h "$SOURCE" ] ; do SOURCE="$(readlink "$SOURCE")"; done
DIR="$( cd -P "$( dirname "$SOURCE" )/../" && pwd )"

# Change into that directory
cd "$DIR"

if [ -z "${XC_ARCH}" ];
then
    XC_ARCH=$(go env GOARCH)
fi

if [ -z "${XC_OS}" ];
then
    XC_OS=$(go env GOOS)
fi

binaryDir="bin/${PNAME}/"$XC_OS"_"$XC_ARCH

# if [ -d "$binaryDir" ]; then
#     echo "Binary does not exist for OS ${XC_OS} and ARCH ${XC_ARCH}"
#     exit 1
# fi

binary="${binaryDir}/${CTLNAME}"

echo $CLTNAME

cp ./$binary .

docker build -t ${STALE_STS_PVC_CLEANER_IMAGE_TAG} -f Dockerfile . --no-cache
