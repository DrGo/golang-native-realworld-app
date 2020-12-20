EXE=rwb
VERSION=0.0.1
PRIV_VERSION=$(shell git describe --tags --abbrev=0)
BUILD := $(shell git describe --always --long --dirty)
BUILD_DATE=$(shell date)
DIST=dist
README=README.md

LDFLAGS=-ldflags "-X main.version=${VERSION} -X main.build=${BUILD}"

#default command
build:
	go build -race ${LDFLAGS} -o ${EXE}

buildprod: 
	go build ${LDFLAGS} -o ${EXE}

# Windows build
buildwin:
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build  -race ${LDFLAGS} -o ${EXE}.exe	

buildwinprod:
	GOOS=windows GOARCH=amd64 go build  ${LDFLAGS} -o ${DIST}/${EXE}.exe	

bench:
	wrk -c 80 -d 5  http://localhost:8080/api/articles

doc: 
#update README with version and build number in line 4 only and only if it had Version
	sed -i '' '4s/.*Version.*/Version ${VERSION} build ${BUILD} on ${BUILD_DATE}/' '${README}'
#update README with the content of the usage.go file
	sed -i 'bak' "/### How to use/,\$$d"  '${README}'
	echo '### How to use' >> '${README}'
	echo '```' >> '${README}'	
	sed -e '1,/Usage/d' usage.go >> '${README}'
	echo '```' >> '${README}'	

cover:
	cd ..
	go test -cover -coverprofile=c.out  
	go tool cover -html=c.out -o coverage.html 
	rm c.out	

showcover:	
	go test -coverprofile=c.out && go tool cover -html=c.out

graph:
	godepgraph "github.com/drgo/rosewood/cmd"  | dot -Tpng -o godepgraph.png	

initdist:
	rm -rf ${DIST}
	mkdir ${DIST}
	cp vdec.css ${DIST}/carpenter.css

finalizedist:
	cp README.md  ${DIST}/README.md
	cd ${DIST}; zip -r dist.zip .	

config:
	echo ${EXE} ${VERSION}-${BUILD} ${PRIV_VERSION} ${DIST}

changelog:
	git log ${PRIV_VERSION}..HEAD --pretty="%s (%an)" --no-merges > ${DIST}/changelog

release: initdist changelog buildwinprod doc finalizedist
	
.PHONY: build buildwin buildprod buildwinprod clean install doc testcover graph release initdist changelog zipdist