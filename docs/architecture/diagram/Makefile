PYTHON_VERSION := 3.9.18
PYENV := $(shell command -v pyenv 2>/dev/null || echo "pyenv not found")
VENV_DIR := .venv

.PHONY: setup
setup: venv install

.PHONY: venv
venv:
	@echo "Creating virtual environment with pyenv..."
	$(PYENV) install -s $(PYTHON_VERSION)
	$(PYENV) local $(PYTHON_VERSION)
	$(PYENV) exec python -m venv $(VENV_DIR)
	@echo "Virtual environment created at $(VENV_DIR)"

.PHONY: install
install: venv
	@echo "Activating virtual environment and installing requirements..."
	. $(VENV_DIR)/bin/activate && pip install -U pip setuptools && pip install -r requirements.txt
	@echo "Requirements installed successfully"

.PHONY: clean
clean:
	@echo "Removing virtual environment..."
	rm -rf $(VENV_DIR)
	@echo "Virtual environment removed"