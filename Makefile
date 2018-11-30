# Default recipes for subdirs
RECIPES := test build
$(RECIPES):
	make -C cgo $@
	make -C cmd $@
