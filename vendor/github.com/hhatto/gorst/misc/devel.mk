RST=./cmd/gorst/gorst

#
# development utilities
#
gofmt:
	rc ./misc/gofmt.rc

diff: ,,c
	tkdiff $< parser.leg

,,c:	,,pmd/markdown_parser.leg
	sed -f misc/c2go.sed < $< > $@

orig-c-src/markdown_parser.leg: orig-c-src


bmprepare:
	rc ./misc/bmprepare.rc

benchmark: cmd m ,,pmd ,,prevmd
	rc ./misc/benchmark.rc


#
# pprof
#
pprof: cmd m
	$(RST) -cpuprofile /tmp/md.prof <m > /tmp/,,md.out
	@echo go tool pprof \'--nodefraction=0.1\' $(RST) /tmp/md.prof
	@echo go tool pprof $(RST) /tmp/md.prof

.PHONY:\
	diff\
	gofmt\
	pprof\
