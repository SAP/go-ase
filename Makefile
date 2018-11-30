# Default recipes for subdirs
RECIPES := test build
$(RECIPES):
	make -C driver $@
	make -C cmd $@
