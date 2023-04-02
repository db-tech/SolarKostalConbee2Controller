package ini

import (
	"fmt"
	"gopkg.in/ini.v1"
)

func LoadPropertiesFromFile(filepath string) (map[string]string, error) {
	// Load the INI file
	cfg, err := ini.Load(filepath)
	if err != nil {
		return nil, fmt.Errorf("error loading INI file: %v", err)
	}

	// Create a map to store the properties
	properties := make(map[string]string)

	// Loop through the sections and keys in the INI file and add them to the map
	for _, section := range cfg.Sections() {
		for _, key := range section.Keys() {
			properties[key.Name()] = key.Value()
		}
	}

	return properties, nil
}

func SavePropertiesToFile(filepath string, properties map[string]string) error {
	// Open the INI file for writing
	cfg, err := ini.Load(filepath)
	if err != nil {
		cfg = ini.Empty()
	}

	// Loop through the properties and set them in the INI file
	for key, value := range properties {
		cfg.Section("").Key(key).SetValue(value)
	}

	// Save the INI file
	err = cfg.SaveTo(filepath)
	if err != nil {
		return fmt.Errorf("error saving INI file: %v", err)
	}

	return nil
}
