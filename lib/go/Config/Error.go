package Config

import "fmt"

const (
	PATH_REQUIRED_TEMPLATE      = "%s path required."
	PATH_DOESNT_EXISTS_TEMPLATE = "%s path doesn't exists."
	INVALID_VALUE_TEMPLATE      = "Invalid value %v for %s."
	REPROCESS_TTL_TEMPLATE      = "Metric %s reprocess of %d days exceeds its ttl."
)

func PathRequired(pathType string) error {
	return fmt.Errorf(PATH_REQUIRED_TEMPLATE, pathType)
}

func PathDoesNotExists(path string) error {
	return fmt.Errorf(PATH_DOESNT_EXISTS_TEMPLATE, path)
}

func InvalidValue(field string, value any) error {
	return fmt.Errorf(INVALID_VALUE_TEMPLATE, value, field)
}

func ReprocessExceedsTTL(id string, days int) error {
	return fmt.Errorf(REPROCESS_TTL_TEMPLATE, id, days)
}
