# GIT_COMMIT and GIT_TAG borrowed from ergo open source project:
# https://raw.githubusercontent.com/ergochat/ergo/master/Makefile

GIT_COMMIT := $(shell git rev-parse HEAD 2> /dev/null)
GIT_TAG := $(shell git tag --points-at HEAD 2> /dev/null | head -n 1)

all:
	go build -ldflags="-X main.Commit=$(GIT_COMMIT) -X main.Version=$(GIT_TAG)"
clean:
	rm -f ./wvlist
debug:
	./wvlist -d -t 7070 -p 6060 run
lint:
	npx htmlhint -r "tagname-lowercase,attr-lowercase,attr-value-double-quotes,doctype-first,doctype-html5,tag-pair,spec-char-escape,id-unique,src-not-empty,attr-no-duplication,title-require,attr-unsafe-chars,html-lang-require,id-unique,tag-pair,tagname-lowercase,tagname-specialchars,attr-no-unnecessary-whitespace,empty-tag-not-self-closed"
