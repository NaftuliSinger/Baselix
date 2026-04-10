package types

type UniqueDuplicateValueError struct {
	FieldName string
	Value     string
}

func (e *UniqueDuplicateValueError) Error() string {
	return "duplicate value: '" + e.Value + "' for unique field: " + e.FieldName
}

type TableAlreadyExistsError struct {
	TableName string
}

func (e *TableAlreadyExistsError) Error() string {
	return "table '" + e.TableName + "' already exists"
}

type ResrvedFieldError struct {
	FieldName string
}

func (e *ResrvedFieldError) Error() string {
	return "cannot use reserved field: " + e.FieldName
}
