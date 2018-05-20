package main

import (
	"encoding/json"
	"fmt"
	"strconv"
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

func parseApivalidatorTag(tag string) ([]string, error) {
	if tag == "" {
		return nil, fmt.Errorf("Empty tag, nothing to")
	}
	splitted := strings.Split(tag, ":")
	if splitted[0] != "apivalidator" {
		return nil, fmt.Errorf("Unknown tag: %s", splitted[0])
	}

	rules := strings.Split(strings.Trim(splitted[1], `"`), ",")

	return rules, nil
}

func parseApivalidatorInt(tag string) (*ApivalidatorInt, error) {
	rules, err := parseApivalidatorTag(tag)
	if err != nil {
		return nil, err
	}

	intRules := &ApivalidatorInt{}

	for _, r := range rules {
		if r == "" {
			return nil, fmt.Errorf("parseApivalidatorInt: Empty rule")
		}

		if r == "required" {
			intRules.Required = true
			continue
		}

		if strings.Contains(r, "paramname") {
			splitted := strings.Split(r, "=")
			if len(splitted) == 0 || len(splitted) > 2 {
				return nil, fmt.Errorf("parseApivalidatorInt: invalid `paramname` declaration")
			}

			intRules.ParamName = splitted[1]
			continue
		}

		if strings.Contains(r, "default") {
			splitted := strings.Split(r, "=")
			if len(splitted) == 0 || len(splitted) > 2 {
				return nil, fmt.Errorf("parseApivalidatorInt: invalid `default` declaration")
			}

			intRules.Default, err = strconv.Atoi(splitted[1])
			if err != nil {
				return nil, err
			}
			continue
		}

		if strings.Contains(r, "min") {
			splitted := strings.Split(r, "=")
			if len(splitted) == 0 || len(splitted) > 2 {
				return nil, fmt.Errorf("parseApivalidatorInt: invalid `min` declaration")
			}

			intRules.Min, err = strconv.Atoi(splitted[1])
			if err != nil {
				return nil, err
			}
			continue
		}

		if strings.Contains(r, "max") {
			splitted := strings.Split(r, "=")
			if len(splitted) == 0 || len(splitted) > 2 {
				return nil, fmt.Errorf("parseApivalidatorInt: invalid `max` declaration")
			}

			intRules.Max, err = strconv.Atoi(splitted[1])
			if err != nil {
				return nil, err
			}
		}
	}

	return intRules, nil
}

func parseApivalidatorString(tag string) (*ApiValidatorString, error) {
	rules, err := parseApivalidatorTag(tag)
	if err != nil {
		return nil, err
	}

	stringRules := &ApiValidatorString{}

	for _, r := range rules {
		if r == "" {
			return nil, fmt.Errorf("parseApivalidatorString: Empty rule")
		}

		if r == "required" {
			stringRules.Required = true
			continue
		}

		if strings.Contains(r, "paramname") {
			splitted := strings.Split(r, "=")
			if len(splitted) == 0 || len(splitted) > 2 {
				return nil, fmt.Errorf("parseApivalidatorString: invalid `paramname` declaration")
			}

			stringRules.ParamName = splitted[1]
			continue
		}

		if strings.Contains(r, "default") {
			splitted := strings.Split(r, "=")
			if len(splitted) == 0 || len(splitted) > 2 {
				return nil, fmt.Errorf("parseApivalidatorString: invalid `default` declaration")
			}

			stringRules.Default = splitted[1]
			continue
		}

		if strings.HasPrefix(r, "min") {
			splitted := strings.Split(r, "=")
			if len(splitted) == 0 || len(splitted) > 2 {
				return nil, fmt.Errorf("parseApivalidatorString: invalid `min` declaration")
			}

			stringRules.Min, err = strconv.Atoi(splitted[1])
			if err != nil {
				return nil, err
			}
			continue
		}

		if strings.Contains(r, "max") {
			splitted := strings.Split(r, "=")
			if len(splitted) == 0 || len(splitted) > 2 {
				return nil, fmt.Errorf("parseApivalidatorString: invalid `max` declaration")
			}

			stringRules.Max, err = strconv.Atoi(splitted[1])
			if err != nil {
				return nil, err
			}
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

			stringRules.Enum = roles
		}
	}

	return stringRules, nil
}
