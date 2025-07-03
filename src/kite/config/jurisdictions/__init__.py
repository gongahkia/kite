import os
import yaml

def get_jurisdiction_config(jurisdiction):
    base_dir = os.path.dirname(os.path.abspath(__file__))
    config_file = f"{jurisdiction.lower()}_config.yaml"
    config_path = os.path.join(base_dir, config_file)
    if not os.path.exists(config_path):
        raise FileNotFoundError(f"Jurisdiction config not found: {config_path}")
    with open(config_path, "r", encoding="utf-8") as f:
        return yaml.safe_load(f)