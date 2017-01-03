# This how we want to name the binary output
BINARY=pingo.exe

# These are the values we want to pass for VERSION and BUILD
VERSION=1.3.0
BUILD=`git rev-parse HEAD`

GOOS=windows
GOARCH=amd64

# Setup the -ldflags option for go build here, interpolate the variable values
LDFLAGS=-ldflags "-X main.Version=${VERSION} -X main.Build=${BUILD}"

# Builds the project
build:
	go build -v ${LDFLAGS} -o ${BINARY} .

# Installs our project: copies binaries
install:
	go install ${LDFLAGS}

# Cleans our project: deletes binaries
clean:
	if [ -f ${BINARY} ] ; then rm ${BINARY} ; fi

.PHONY: clean install
