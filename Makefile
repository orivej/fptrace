BIN_TARGETS = depgrapher $(TRACEE)
TEST_TARGETS = $(TESTCMD)
TEST_TEMPS = a b c
OBJECT_FILES = */*.o 

TRACEE = tracee/tracee
TESTCMD = testcmd/testcmd

default: compile

clean:
	rm -f $(BIN_TARGETS) $(TEST_TARGETS) $(TEST_TEMPS) $(OBJECT_FILES)

compile: $(BIN_TARGETS)

test: $(BIN_TARGETS) $(TEST_TARGETS)
	./depgrapher -tracee $(TRACEE) -t /dev/null -d /dev/stdout $(TESTCMD)

install: $(BIN_TARGETS)
	mkdir -p $(DESTDIR)
	cp $(BIN_TARGETS) $(DESTDIR)

depgrapher:
	go build -o $@
