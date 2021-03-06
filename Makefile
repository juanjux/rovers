# Package configuration
PROJECT = rovers
COMMANDS = .

GO_BUILD_ENV = CGO_ENABLED=0

# Including ci Makefile
CI_REPOSITORY ?= https://github.com/src-d/ci.git
CI_PATH ?= .ci
CI_VERSION ?= v1

PKG_OS = linux darwin windows

MAKEFILE := $(CI_PATH)/Makefile.main
$(MAKEFILE):
	git clone --quiet --branch $(CI_VERSION) --depth 1 $(CI_REPOSITORY) $(CI_PATH);

-include $(MAKEFILE)
