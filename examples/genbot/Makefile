CONFIG_DIR = $(HOME)/botmand-engines
VENV_DIR = $(CONFIG_DIR)/genbot-venv

all:
	@echo "Run \"make install\" to set up a virtualenv and install the genbot engine"

install: $(CONFIG_DIR) $(VENV_DIR) $(CONFIG_DIR)/genbot.yaml $(CONFIG_DIR)/genbot.py
	@echo
	@echo "Remember to set the OpenAI API key in $(CONFIG_DIR)/genbot.yaml"
	@echo

$(VENV_DIR):
	python3 -m venv $(VENV_DIR)
	$(VENV_DIR)/bin/pip install -r requirements.txt

$(CONFIG_DIR)/genbot.yaml: genbot.yaml.tmpl
	sed -e 's|@CONFIG_DIR@|$(CONFIG_DIR)|' -e 's|@VENV_DIR@|$(VENV_DIR)|' $< > $@

$(CONFIG_DIR)/genbot.py: genbot.py.tmpl
	sed -e 's|@CONFIG_DIR@|$(CONFIG_DIR)|' -e 's|@VENV_DIR@|$(VENV_DIR)|' $< > $@
	chmod 755 $@

$(CONFIG_DIR):
	mkdir -p $@

clean:
	$(RM) $(CONFIG_DIR)/genbot.yaml $(CONFIG_DIR)/genbot.py
	$(RM) -r $(VENV_DIR)
