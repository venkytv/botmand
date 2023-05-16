CONFIG_DIR = $(HOME)/botters-engines

all:
	@echo Run \"make install\" to install the mentionbot engine

install: $(CONFIG_DIR) $(CONFIG_DIR)/mentionbot.yaml $(CONFIG_DIR)/mentionbot.sh

$(CONFIG_DIR)/mentionbot.yaml: mentionbot.yaml.tmpl
	sed -e 's|@CONFIG_DIR@|$(CONFIG_DIR)|' $< > $@

$(CONFIG_DIR)/mentionbot.sh: mentionbot.sh
	cp $< $@
	chmod 755 $@

$(CONFIG_DIR):
	mkdir -p $@

clean:
	$(RM) $(CONFIG_DIR)/mentionbot.yaml $(CONFIG_DIR)/mentionbot.sh
