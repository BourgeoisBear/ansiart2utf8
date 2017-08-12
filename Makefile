# := EXPAND ON DEFINITION
#  = EXPAND ON USE
APPNAME     := ansiart2utf8

# FOR LIVE TESTING (REQUIRES ag (siver-searcher) AND entr)
# ag -l --go | entr -s 'make buildtest'

# find -type f -name '*.ans' -exec ../../ansiart2utf8 -d -f {} \;

default: build

# NOTE: COMMANDS ON DIFFERENT LINES ARE RUN IN DIFFERENT SHELLS
buildtest:
	reset ; go build -v -o $(APPNAME) ; if [ $$? -eq 0 ] ; then make test ; fi

test:
	./ansiart2utf8 -d -f ./test_data/ZOMBIE_KILLING.ans
	./ansiart2utf8 -d -f ./test_data/_07_Calendar_2017_July_by_Andy_Herbert.ans
	./ansiart2utf8 -d -f ./test_data/fruit.ans
#	./ansiart2utf8 -d -w 200 -f ./bt-will_be_blocks/WZ\ -\ DJAC.ans
#	./ansiart2utf8 -d -w 220 -f ./bt-will_be_blocks/WZ\ -\ Gord\ Downie.ans | tee ./processed.ans

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