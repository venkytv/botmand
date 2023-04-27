CONFIG_DIR = $(HOME)/teabot-engines
IMAGE_NAME = teabot-gptbot

all:
	@echo "Run \"make install\" to build docker image and install the gptbot engine"

install: $(CONFIG_DIR) build $(CONFIG_DIR)/gptbot.yaml $(CONFIG_DIR)/gptbot.sh
	@echo
	@echo "Remember to set the OpenAI API key in $(CONFIG_DIR)/gptbot.yaml"
	@echo

build:
	docker build -t $(IMAGE_NAME) .

$(CONFIG_DIR)/gptbot.yaml: gptbot.yaml.tmpl
	sed -e 's|@CONFIG_DIR@|$(CONFIG_DIR)|' $< > $@

$(CONFIG_DIR)/gptbot.sh: gptbot.sh.tmpl
	sed -e 's|@IMAGE_NAME@|$(IMAGE_NAME)|' $< > $@
	chmod 755 $@

$(CONFIG_DIR):
	mkdir -p $@

clean:
	$(RM) $(CONFIG_DIR)/gptbot.yaml $(CONFIG_DIR)/gptbot.sh
