# := EXPAND ON DEFINITION
#  = EXPAND ON USE

default: build

test:
	go run ./ansiart2utf8 -d -f ./test_data/ZOMBIE_KILLING.ans
	go run ./ansiart2utf8 -d -f ./test_data/_07_Calendar_2017_July_by_Andy_Herbert.ans
	go run ./ansiart2utf8 -d -f ./test_data/fruit.ans
	#go run ./ansiart2utf8 -d -w 200 -f ./bt-will_be_blocks/WZ\ -\ DJAC.ans
	#go run ./ansiart2utf8 -d -w 220 -f ./bt-will_be_blocks/WZ\ -\ Gord\ Downie.ans | tee ./processed.ans

debug:
	find -type f -iname '*.go' | entr -c -r go test -v

tidy:
	go mod tidy
