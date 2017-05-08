BIN_TARGETS = fptrace $(TRACEE)
TEST_TARGETS = $(TESTCMD) $(SEGFAULT)
TEST_TEMPS = a b c
OBJECT_FILES = */*.o

TRACEE = ./_fptracee
TESTCMD = testcmd/testcmd
SEGFAULT = testcmd/segfault

DESTDIR ?= $(shell echo "$${GOBIN:-$${GOPATH/:*/}/bin}")

default: compile

clean:
	rm -f $(BIN_TARGETS) $(TEST_TARGETS) $(TEST_TEMPS) $(OBJECT_FILES)

compile: $(BIN_TARGETS)

test: $(BIN_TARGETS) $(TEST_TARGETS)
	./fptrace -tracee $(TRACEE) -d /dev/stdout $(TESTCMD)
	./fptrace -tracee $(TRACEE) -t /dev/stdout $(SEGFAULT)

install: $(BIN_TARGETS)
	mkdir -p $(DESTDIR)
	cp $(BIN_TARGETS) $(DESTDIR)

fptrace: *.go
	go build -o $@
