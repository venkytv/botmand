CONFIG_DIR = $(HOME)/botmand-engines

all:
	@echo Run \"make install\" to install the cuckoobot engine

install: $(CONFIG_DIR) $(CONFIG_DIR)/cuckoobot.yaml $(CONFIG_DIR)/cuckoobot.sh

$(CONFIG_DIR)/cuckoobot.yaml: cuckoobot.yaml.tmpl
	sed -e 's|@CONFIG_DIR@|$(CONFIG_DIR)|' $< > $@

$(CONFIG_DIR)/cuckoobot.sh: cuckoobot.sh
	cp $< $@
	chmod 755 $@

$(CONFIG_DIR):
	mkdir -p $@

clean:
	$(RM) $(CONFIG_DIR)/cuckoobot.yaml $(CONFIG_DIR)/cuckoobot.sh
