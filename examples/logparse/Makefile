CONFIG_DIR = $(HOME)/botmand-engines
VENV_DIR = $(CONFIG_DIR)/logparse-venv

all:
	@echo "Run \"make install\" to set up a virtualenv and install the logparse engine"

install: $(CONFIG_DIR) $(VENV_DIR) $(CONFIG_DIR)/logparse.yaml $(CONFIG_DIR)/logparse.py
	@echo
	@echo "Remember to set the OpenAI API key in $(CONFIG_DIR)/logparse.yaml"
	@echo

$(VENV_DIR):
	python3 -m venv $(VENV_DIR)
	$(VENV_DIR)/bin/pip install -r requirements.txt

$(CONFIG_DIR)/logparse.yaml: logparse.yaml.tmpl
	sed -e 's|@CONFIG_DIR@|$(CONFIG_DIR)|' -e 's|@VENV_DIR@|$(VENV_DIR)|' $< > $@

$(CONFIG_DIR)/logparse.py: logparse.py.tmpl
	sed -e 's|@CONFIG_DIR@|$(CONFIG_DIR)|' -e 's|@VENV_DIR@|$(VENV_DIR)|' $< > $@
	chmod 755 $@

$(CONFIG_DIR):
	mkdir -p $@

clean:
	$(RM) $(CONFIG_DIR)/logparse.yaml $(CONFIG_DIR)/logparse.py
	$(RM) -r $(VENV_DIR)
