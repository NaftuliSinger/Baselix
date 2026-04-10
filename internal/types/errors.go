package types

type UniqueDuplicateValueError struct {
	FieldName string
	Value     string
}

func (e *UniqueDuplicateValueError) Error() string {
	return "duplicate value: \"" + e.Value + "\" for unique field: " + e.FieldName
}

type ResrvedFieldError struct {
	FieldName string
}

func (e *ResrvedFieldError) Error() string {
	return "cannot use reserved field: " + e.FieldName
}
