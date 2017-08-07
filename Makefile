# := EXPAND ON DEFINITION
#  = EXPAND ON USE
APPNAME     := ansiart2utf8

# TODO: INCLUDE TEST IMAGES IN GITHUB
# TODO: NEW DEMO IMAGE IN MARKDOWN FILE

# FOR LIVE TESTING (REQUIRES ag (siver-searcher) AND entr)
# ag -l --go | entr -s 'make buildtest'

# find -type f -name 'l*.ans' -exec ../../ansiart2utf8 -d -f {} \;

# ls TEST.ans | entr -c -s 'cat ./TEST.ans'

default: build

# NOTE: COMMANDS ON DIFFERENT LINES ARE RUN IN DIFFERENT SHELLS
buildtest:
	reset ; go build -v -o $(APPNAME) ; if [ $$? -eq 0 ] ; then make test ; fi

test:
	./ansiart2utf8 -d -f ./bt-will_be_blocks/ZOMBIE_KILLING.ans
#	./ansiart2utf8 -d -f ./bt-will_be_blocks/_07_Calendar_2017_July_by_Andy_Herbert.ans
	./ansiart2utf8 -d -f ./textfiles/artwork/fruit.ans
#  ./ansiart2utf8 -d -w 200 -f ./bt-will_be_blocks/WZ\ -\ DJAC.ans

debug:
	ag -l --go | entr -s 'make run'

fbuild: clean
	go build -a -v -o $(APPNAME)

build: clean
	go build -v -o $(APPNAME)

clean:
	rm -f $(APPNAME)

depend:
	go get -d -v ./...