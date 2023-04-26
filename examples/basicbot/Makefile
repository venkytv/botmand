CONFIG_DIR = $(HOME)/teabot-engines

all:
	@echo Run \"make install\" to install the basicbot engine

install: $(CONFIG_DIR) $(CONFIG_DIR)/basicbot.yaml $(CONFIG_DIR)/basicbot.sh

$(CONFIG_DIR)/basicbot.yaml: basicbot.yaml.tmpl
	sed -e 's|@CONFIG_DIR@|$(CONFIG_DIR)|' $< > $@

$(CONFIG_DIR)/basicbot.sh: basicbot.sh
	cp $< $@
	chmod 755 $@

$(CONFIG_DIR):
	mkdir -p $@
