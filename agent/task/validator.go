package task

import (
	"fmt"
	"reflect"
	"regexp"
)

type validator func(message) error

func (t *Task) newRequiredValidtor(conf validatorConf) validator {
	if len(conf.Column) == 0 {
		t.configureFatal("required validation", "column")
	}
	return func(msg message) error {
		if _, ok := msg[conf.Column]; !ok {
			return fmt.Errorf("column %s is required", conf.Column)
		}
		return nil
	}
}

func (t *Task) newTypeValidtor(conf validatorConf) validator {
	if len(conf.Column) == 0 {
		t.configureFatal("type validation", "column")
	}
	var kind reflect.Kind
	switch conf.Type {
	case "string":
		kind = reflect.String
	case "int":
		kind = reflect.Int
	case "float":
		kind = reflect.Float64
	default:
		t.logger.Fatal("unsupported validation type %s", conf.Type)
	}
	return func(msg message) error {
		v, ok := msg[conf.Column]
		if !ok {
			return fmt.Errorf("column %s type validation failed", conf.Column)
		}
		if reflect.TypeOf(v).Kind() != kind {
			return fmt.Errorf("column %s type validation failed", conf.Column)
		}
		return nil
	}
}

func (t *Task) newValueValidtor(conf validatorConf) validator {
	if len(conf.Column) == 0 {
		t.configureFatal("value validation", "column")
	}
	if len(conf.Value) == 0 {
		t.configureFatal("value validation", "value")
	}
	return func(msg message) error {
		v, ok := msg[conf.Column].(string)
		if !ok {
			return fmt.Errorf("column %s value validation failed", conf.Column)
		}
		if v != conf.Value {
			return fmt.Errorf("column %s value validation failed", conf.Column)
		}
		return nil
	}
}

func (t *Task) newNumberValidtor(conf validatorConf) validator {
	if len(conf.Column) == 0 {
		t.configureFatal("value validation", "column")
	}
	return func(msg message) error {
		v, ok := msg[conf.Column].(int)
		if !ok {
			return fmt.Errorf("column %s number validation failed", conf.Column)
		}
		if v != conf.Number {
			return fmt.Errorf("column %s number validation failed", conf.Column)
		}
		return nil
	}
}

func (t *Task) newMaxValueValidtor(conf validatorConf) validator {
	if len(conf.Column) == 0 {
		t.configureFatal("maxvalue validation", "column")
	}
	return func(msg message) error {
		v, ok := msg[conf.Column].(int)
		if !ok {
			return fmt.Errorf("column %s maxvalue validation failed", conf.Column)
		}
		if v > conf.Number {
			return fmt.Errorf("column %s maxvalue validation failed", conf.Column)
		}
		return nil
	}
}

func (t *Task) newMinValueValidtor(conf validatorConf) validator {
	if len(conf.Column) == 0 {
		t.configureFatal("minvalue validation", "column")
	}
	return func(msg message) error {
		v, ok := msg[conf.Column].(int)
		if !ok {
			return fmt.Errorf("column %s minvalue validation failed", conf.Column)
		}
		if v < conf.Number {
			return fmt.Errorf("column %s minvalue validation failed", conf.Column)
		}
		return nil
	}
}

func (t *Task) newMaxLengthValidtor(conf validatorConf) validator {
	if len(conf.Column) == 0 {
		t.configureFatal("maxlength validation", "column")
	}
	return func(msg message) error {
		v, ok := msg[conf.Column].(string)
		if !ok {
			return fmt.Errorf("column %s maxlength validation failed", conf.Column)
		}
		if len(v) > conf.Number {
			return fmt.Errorf("column %s maxlength validation failed", conf.Column)
		}
		return nil
	}
}

func (t *Task) newMinLengthValidtor(conf validatorConf) validator {
	if len(conf.Column) == 0 {
		t.configureFatal("minlength validation", "column")
	}
	return func(msg message) error {
		v, ok := msg[conf.Column].(string)
		if !ok {
			return fmt.Errorf("column %s minlength validation failed", conf.Column)
		}
		if len(v) < conf.Number {
			return fmt.Errorf("column %s minlength validation failed", conf.Column)
		}
		return nil
	}
}

func (t *Task) newRegexValidtor(conf validatorConf) validator {
	if len(conf.Column) == 0 {
		t.configureFatal("regex validation", "column")
	}
	if len(conf.Regex) == 0 {
		t.configureFatal("regex validation", "regex")
	}
	cmp, err := regexp.Compile(conf.Regex)
	if err != nil {
		t.logger.Fatal("invalid configuration `regex`")
	}
	return func(msg message) error {
		v, ok := msg[conf.Column].(string)
		if !ok {
			return fmt.Errorf("column %s regex validation failed", conf.Column)
		}
		if !cmp.MatchString(v) {
			return fmt.Errorf("column %s regex validation failed", conf.Column)
		}
		return nil
	}
}

func (t *Task) initValidator(conf validatorConf) validator {
	switch conf.Mode {
	case "required":
		return t.newRequiredValidtor(conf)
	case "type":
		return t.newTypeValidtor(conf)
	case "value":
		return t.newValueValidtor(conf)
	case "number":
		return t.newNumberValidtor(conf)
	case "maxvalue":
		return t.newMaxValueValidtor(conf)
	case "minvalue":
		return t.newMinValueValidtor(conf)
	case "maxlength":
		return t.newMaxLengthValidtor(conf)
	case "minlength":
		return t.newMinLengthValidtor(conf)
	case "regex":
		return t.newRegexValidtor(conf)
	default:
		t.logger.Fatalf("unsupported validation mode `%s`", conf.Mode)
		return nil
	}
}

func (t *Task) initValidators(validatorsConf []validatorConf) (validators []validator) {
	for _, conf := range validatorsConf {
		v := t.initValidator(conf)
		validators = append(validators, v)
	}
	return validators
}

func (t *Task) initGlobalValidators(validatorsConf []validatorConf) {
	vs := t.initValidators(validatorsConf)
	if &t.processor != nil {
		func(processor process) {
			t.processor = func(msg message) (message, error) {
				msg, err := processor(msg)
				if err != nil {
					return msg, err
				}
				for _, v := range vs {
					if err := v(msg); err != nil {
						return msg, err
					}
				}
				return msg, nil
			}
		}(t.processor)
	} else {
		t.processor = func(msg message) (message, error) {
			for _, v := range vs {
				if err := v(msg); err != nil {
					return msg, err
				}
			}
			return msg, nil
		}
	}
}
