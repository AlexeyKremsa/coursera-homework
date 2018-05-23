package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

func parseReceiverType(name string) string {
	splittedArr := strings.Split(name, " ")
	if len(splittedArr) == 0 {
		return ""
	}

	lastElem := splittedArr[len(splittedArr)-1]
	lastElem = strings.TrimRight(lastElem, "}")

	return lastElem
}

func parseApigenComment(comment string) (*ApigenComment, error) {
	start := strings.Index(comment, "{")
	end := strings.Index(comment, "}")
	finalStr := comment[start : end+1]

	tag := strings.TrimSpace(comment[:start])
	if tag != "apigen:api" {
		return nil, fmt.Errorf("unknown tag: %s", tag)
	}

	apigen := &ApigenComment{}
	err := json.Unmarshal([]byte(finalStr), apigen)
	if err != nil {
		return nil, err
	}

	return apigen, nil
}

func getApivalidatorTag(tag string) ([]string, error) {
	if tag == "" {
		return nil, fmt.Errorf("Empty tag, nothing to parse")
	}
	splitted := strings.Split(tag, ":")
	if splitted[0] != "apivalidator" {
		return nil, fmt.Errorf("Unknown tag: %s", splitted[0])
	}

	rules := strings.Split(strings.Trim(splitted[1], `"`), ",")

	return rules, nil
}

func parseApivalidatorTags(fieldType string, tag string) (*ApiValidatorTags, error) {
	rules, err := getApivalidatorTag(tag)
	if err != nil {
		return nil, err
	}

	tags := &ApiValidatorTags{}

	switch fieldType {
	case "int":
		for _, r := range rules {
			if r == "" {
				return nil, fmt.Errorf("parseApivalidatorInt: Empty rule")
			}

			if r == "required" {
				tags.Required = true
				continue
			}

			if strings.Contains(r, "paramname") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorInt: invalid `paramname` declaration")
				}

				tags.ParamName = splitted[1]
				continue
			}

			if strings.Contains(r, "default") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorInt: invalid `default` declaration")
				}

				tags.DefaultInt = splitted[1]
				continue
			}

			if strings.Contains(r, "min") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorInt: invalid `min` declaration")
				}

				tags.Min = splitted[1]
				continue
			}

			if strings.Contains(r, "max") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorInt: invalid `max` declaration")
				}

				tags.Max = splitted[1]
			}
		}

		return tags, nil

	case "string":
		for _, r := range rules {
			if r == "" {
				return nil, fmt.Errorf("parseApivalidatorString: Empty rule")
			}

			if r == "required" {
				tags.Required = true
				continue
			}

			if strings.Contains(r, "paramname") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorString: invalid `paramname` declaration")
				}

				tags.ParamName = splitted[1]
				continue
			}

			if strings.Contains(r, "default") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorString: invalid `default` declaration")
				}

				tags.DefaultString = splitted[1]
				continue
			}

			if strings.HasPrefix(r, "min") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorString: invalid `min` declaration")
				}

				tags.Min = splitted[1]
				continue
			}

			if strings.Contains(r, "max") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorString: invalid `max` declaration")
				}

				tags.Max = splitted[1]
			}

			if strings.Contains(r, "enum") {
				splitted := strings.Split(r, "=")
				if len(splitted) == 0 || len(splitted) > 2 {
					return nil, fmt.Errorf("parseApivalidatorString: invalid `max` declaration")
				}

				roles := strings.Split(splitted[1], "|")
				if len(roles) == 0 || len(roles) > 3 {
					return nil, fmt.Errorf("parseApivalidatorString: invalid enum declaration")
				}

				tags.Enum = roles
			}
		}

		return tags, nil

	default:
		return nil, fmt.Errorf("unsupported type: %s", fieldType)
	}
}
