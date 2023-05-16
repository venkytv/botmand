CONFIG_DIR = $(HOME)/botters-engines

all:
	@echo Run \"make install\" to install the gamebot engine

install: $(CONFIG_DIR) $(CONFIG_DIR)/gamebot.yaml $(CONFIG_DIR)/gamebot.sh

$(CONFIG_DIR)/gamebot.yaml: gamebot.yaml.tmpl
	sed -e 's|@CONFIG_DIR@|$(CONFIG_DIR)|' $< > $@

$(CONFIG_DIR)/gamebot.sh: gamebot.sh
	cp $< $@
	chmod 755 $@

$(CONFIG_DIR):
	mkdir -p $@

clean:
	$(RM) $(CONFIG_DIR)/gamebot.yaml $(CONFIG_DIR)/gamebot.sh
