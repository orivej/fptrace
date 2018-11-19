BIN_TARGETS = fptrace $(TRACEE)
TEST_TARGETS = $(TESTCMD) $(SEGFAULT)
TEMPS = a b c *.h */*.o

TRACEE = ./_fptracee
TESTCMD = testcmd/testcmd
SEGFAULT = testcmd/segfault

DESTDIR ?= $(shell bash -c 'GOPATH=$$(go env GOPATH); echo $${GOBIN:-$${GOPATH/:*/}/bin}')

default: compile

clean:
	rm -f $(BIN_TARGETS) $(TEST_TARGETS) $(TEMPS)

compile: $(BIN_TARGETS)

test: $(BIN_TARGETS) $(TEST_TARGETS)
	./fptrace -tracee $(TRACEE) -d /dev/stdout $(TESTCMD)
	./fptrace -tracee $(TRACEE) -d /dev/stdout -seccomp=false $(TESTCMD)
	! ./fptrace -tracee $(TRACEE) -t /dev/stdout $(SEGFAULT)

install: $(BIN_TARGETS)
	mkdir -p $(DESTDIR)
	cp $(BIN_TARGETS) $(DESTDIR)

fptrace: *.go
	go build -o $@

$(TRACEE): seccomp.h

seccomp.h: seccomp.go
	go run seccomp.go
