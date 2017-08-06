# := EXPAND ON DEFINITION
#  = EXPAND ON USE
APPNAME     := ansiart2utf8

#ag -l --go | entr -c -s 'make buildtest'

default: build

# NOTE: COMMANDS ON DIFFERENT LINES ARE RUN IN DIFFERENT SHELLS
buildtest:
	reset ; go build -v -o $(APPNAME) ; if [ $$? -eq 0 ] ; then make test ; fi

test:
	./ansiart2utf8 -d -f ./bt-will_be_blocks/ZOMBIE_KILLING.ans
	./ansiart2utf8 -d -f ./bt-will_be_blocks/_07_Calendar_2017_July_by_Andy_Herbert.ans
	./ansiart2utf8 -d -f ./textfiles/fruit.ans

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