CONFIG_DIR = $(HOME)/teabot-engines

all:
	@echo Run \"make install\" to install the timebot engine

install: $(CONFIG_DIR) $(CONFIG_DIR)/timebot.yaml $(CONFIG_DIR)/timebot.sh

$(CONFIG_DIR)/timebot.yaml: timebot.yaml.tmpl
	sed -e 's|@CONFIG_DIR@|$(CONFIG_DIR)|' $< > $@

$(CONFIG_DIR)/timebot.sh: timebot.sh
	cp $< $@
	chmod 755 $@

$(CONFIG_DIR):
	mkdir -p $@

clean:
	$(RM) $(CONFIG_DIR)/timebot.yaml $(CONFIG_DIR)/timebot.sh
